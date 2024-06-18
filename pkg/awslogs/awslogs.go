package awslogs

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"lkrouter/config"
	"log"
	"sync"
	"time"
)

const (
	// LogTypes for log messages
	MsgTypeInfo  = "info"
	MsgTypeError = "error"
	MsgTypeWarn  = "warning"
)

type CwlLogMessage struct {
	Func      string `json:"func"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Room      string `json:"room"`
	Uid       string `json:"uid"`
	Timestamp int64  `json:"timestamp"`
}

type QueueItem struct {
	Message   string
	Timestamp int64
}

type CwlLogs struct {
	cwl           *cloudwatchlogs.CloudWatchLogs
	logGroupName  string
	logStreamName string
	sequenceToken string
	queue         []QueueItem
	queueLock     sync.Mutex
	logger        *logrus.Logger
}

var (
	cwlLogs  *CwlLogs
	cwlError error
	cwlOnce  sync.Once
)

func createCloudwatchLogs() *cloudwatchlogs.CloudWatchLogs {
	cfg := config.GetConfig()

	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(cfg.AwsLogRegion),
		},
	})

	if err != nil {
		cwlError = err
	}

	return cloudwatchlogs.New(sess)
}

func createCwl() {
	cfg := config.GetConfig()
	cwl := createCloudwatchLogs()

	cwlLogs = &CwlLogs{
		cwl:           cwl,
		logGroupName:  cfg.AwsLogGroupName,
		logStreamName: "",
		sequenceToken: "",
		queue:         []QueueItem{},
		queueLock:     sync.Mutex{},
		logger:        logrus.New(),
	}

	err := cwlLogs.ensureLogGroupExists(cfg.AwsLogGroupName)
	if err != nil {
		cwlError = err
	}
}

func GetCwl() (*CwlLogs, error) {
	//Perform connection creation operation only once.
	cwlOnce.Do(func() {
		cfg := config.GetConfig()
		createCwl()
		fmt.Println("cwlLogs: ", cfg.AwsLogGroupName)
	})
	return cwlLogs, cwlError
}

// ensureLogGroupExists first checks if the log group exists,
// if it doesn't it will create one.
func (c *CwlLogs) ensureLogGroupExists(name string) error {

	//check if log cwl exists
	if c.cwl == nil {
		c.cwl = createCloudwatchLogs()
	}

	resp, err := c.cwl.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{})
	if err != nil {
		return err
	}

	for _, logGroup := range resp.LogGroups {
		if *logGroup.LogGroupName == name {
			return nil
		}
	}

	_, err = c.cwl.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: &name,
	})
	if err != nil {
		return err
	}

	_, err = c.cwl.PutRetentionPolicy(&cloudwatchlogs.PutRetentionPolicyInput{
		RetentionInDays: aws.Int64(14),
		LogGroupName:    &name,
	})

	return err
}

// createLogStream will make a new logStream with a random uuid as its name.
func (c *CwlLogs) createLogStream() error {
	name := uuid.New().String()

	//check if log cwl exists
	if c.cwl == nil {
		c.cwl = createCloudwatchLogs()
	}

	_, err := c.cwl.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  &c.logGroupName,
		LogStreamName: &name,
	})

	c.logStreamName = name

	return err
}

// processQueue will process the log queue
func (c *CwlLogs) ProcessQueue() {
	for {
		logEvents := c.getLogEventsFromQueue()
		if len(logEvents) > 0 {
			input := c.createLogEventsInput(logEvents)
			err := c.putLogEvents(input)
			if err != nil {
				log.Println(err)
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func (c *CwlLogs) getLogEventsFromQueue() []*cloudwatchlogs.InputLogEvent {

	c.queueLock.Lock()
	defer c.queueLock.Unlock()

	var logEvents []*cloudwatchlogs.InputLogEvent

	if len(c.queue) > 0 {

		for _, item := range c.queue {
			messageItem := item
			timestamp := messageItem.Timestamp
			if timestamp == 0 {
				timestamp = time.Now().UnixNano() / int64(time.Millisecond)
			}

			logEvents = append(logEvents, &cloudwatchlogs.InputLogEvent{
				Message:   &messageItem.Message,
				Timestamp: aws.Int64(timestamp),
			})
		}

		c.queue = []QueueItem{}

	}

	return logEvents

}

func (c *CwlLogs) createLogEventsInput(logEvents []*cloudwatchlogs.InputLogEvent) cloudwatchlogs.PutLogEventsInput {

	input := cloudwatchlogs.PutLogEventsInput{
		LogEvents:    logEvents,
		LogGroupName: &c.logGroupName,
	}

	if c.sequenceToken == "" {
		err := c.createLogStream()
		if err != nil {
			panic(err)
		}
	} else {
		input.SetSequenceToken(c.sequenceToken)
	}

	input.SetLogStreamName(c.logStreamName)

	return input

}

func (c *CwlLogs) putLogEvents(input cloudwatchlogs.PutLogEventsInput) error {

	//check if log cwl exists
	if c.cwl == nil {
		c.cwl = createCloudwatchLogs()
	}

	resp, err := c.cwl.PutLogEvents(&input)
	if err != nil {
		c.logger.Infof("CwlLogs.putLogEvents err: {%s}", err)
		return err
	}
	if resp != nil {
		c.sequenceToken = *resp.NextSequenceToken
	}
	return err
}

func (c *CwlLogs) Add(msg string) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()
	c.queue = append(c.queue, QueueItem{
		Message:   msg,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	})
}

// add stract CwlLogMessage
func (c *CwlLogs) AddSLog(mes CwlLogMessage) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()
	mes.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	jsonMes, err := json.Marshal(mes)
	if err != nil {
		c.queue = append(c.queue, QueueItem{fmt.Sprintf("%+v", mes), mes.Timestamp})
	} else {
		c.queue = append(c.queue, QueueItem{string(jsonMes), mes.Timestamp})
	}
}

func AddSLog(mes map[string]string) {
	cwl, _ := GetCwl()
	if cwl != nil {
		cwl.AddSLog(CwlLogMessage{
			Func:    mes["func"],
			Message: mes["message"],
			Room:    mes["room"],
			Uid:     mes["uid"],
			Type:    mes["type"],
		})
	}
}
