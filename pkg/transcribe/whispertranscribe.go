package transcribe

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"github.com/tcolgate/mp3"
	"io"
	"lkrouter/config"
	gcp2 "lkrouter/pkg/gcp"
	"lkrouter/pkg/mongodb/mrequests"
	"lkrouter/utils"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

const (
	fileSizeLimit = 24.5
	segmentSize   = 24
)

type OpenAiTranscribe struct {
	logger       *logrus.Logger
	speechClient *openai.Client
	lang         string
	room         string
	audioUrl     string
	prompt       string
}

func NewWhisperTranscriber(room string, audioUrl string, lang string, prompt string) *OpenAiTranscribe {
	logger := logrus.New()
	cfg := config.GetConfig()

	client := openai.NewClient(cfg.OpenaiApiKey)

	return &OpenAiTranscribe{
		room:         room,
		audioUrl:     audioUrl,
		logger:       logger,
		speechClient: client,
		lang:         lang,
		prompt:       prompt,
	}
}

func (t *OpenAiTranscribe) MakeRequestToWhisper(filePath string) (*openai.AudioResponse, error) {

	prompt := ""
	if t.prompt != "" {
		prompt = t.prompt
	}

	// Transcribe the audio file
	resp, err := t.speechClient.CreateTranscription(
		context.Background(),
		openai.AudioRequest{
			Model:       openai.Whisper1,
			FilePath:    filePath,
			Format:      openai.AudioResponseFormatVerboseJSON,
			Prompt:      prompt,
			Temperature: 0.0,
		},
	)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return nil, err
	}

	if len(resp.Segments) == 0 {
		t.logger.Info("No segments found")
		return nil, fmt.Errorf("No segments found")
	}

	return &resp, nil
}

func (t *OpenAiTranscribe) MakeWhisperFileTranscribe(filePath string) ([]map[string]interface{}, int32, error) {
	resp, err := t.MakeRequestToWhisper(filePath)
	if err != nil {
		t.logger.Errorf("Failed to transcribe audio file: %v", err)
		return nil, 0, err
	}

	t.logger.Printf("Transcription Results: %+v\n", resp.Segments)
	formatedWordsMap := t.FormatTranscribResult(resp)
	t.logger.Printf("Transcription Results: %+v\n", formatedWordsMap)

	return formatedWordsMap, int32(resp.Duration), nil

}

func (t *OpenAiTranscribe) GetSpeachDuration(filePath string) (float64, error) {
	var dt float64 = 0.0

	r, err := os.Open(filePath)
	if err != nil {
		t.logger.Errorf("Failed to open file: %v", err)
		return 0, err
	}
	defer r.Close()

	d := mp3.NewDecoder(r)
	var f mp3.Frame
	skipped := 0

	for {
		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			t.logger.Errorf("Failed to decode mp3: %v", err)
			return 0, err
		}

		dt = dt + f.Duration().Seconds()
	}

	return dt, nil
}

func (t *OpenAiTranscribe) GetSegmentDurations(filePath string, segmentSize int) (int, int) {
	var dt float64 = 0.0
	var sigmentSize int = 0

	dt, err := t.GetSpeachDuration(filePath)
	if err != nil {
		t.logger.Errorf("Failed to get speech duration: %v", err)
		return 0, 0
	}

	size, err := utils.GetFileSizeInMb(filePath)
	if err != nil {
		t.logger.Errorf("Failed to get file size: %v", err)
		return 0, 0
	}
	koef := math.Ceil(dt / size)

	sigmentSize = int(math.Ceil(float64(segmentSize) * koef))
	countOfSegments := int(math.Ceil(dt / float64(sigmentSize)))

	return sigmentSize, countOfSegments
}

func (t *OpenAiTranscribe) DevideFileByTime(filePath string, fileName string, segmentSize int) int {
	cfg := config.GetConfig()
	segmentDuration, countOfSegments := t.GetSegmentDurations(filePath, segmentSize)
	segmentDurationStr := strconv.Itoa(segmentDuration)
	cmd := exec.Command("ffmpeg", "-i", filePath, "-f", "segment", "-segment_time", segmentDurationStr, "-c", "copy", fileName+"/segment_%03d.mp3")
	cmd.Dir = cfg.TmpFilesPath
	err := cmd.Run()
	if err != nil {
		t.logger.Errorf("Failed to devide file by time: %v", err)
	}

	return countOfSegments
}

func (t *OpenAiTranscribe) WhisperFileTranscribe() {
	mrequests.UpdateTranscribeTextStatus(t.room, "progress")

	gcpService := gcp2.NewService()
	fileUrl, err := gcpService.GetSignedURL(t.audioUrl)
	if err != nil {
		t.logger.Errorf("Failed to get signed url: %v", err)
		return
	}

	filePath, err := t.DownloadAudioFile(fileUrl)
	if err != nil {
		t.logger.Errorf("Failed to download audio file: %v", err)
		mrequests.UpdateTranscribeTextStatus(t.room, "error")
		return
	}

	fileSize, err := utils.GetFileSizeInMb(filePath)
	if err != nil {
		t.logger.Errorf("Error when try to get file size: %v", err)
		return
	}

	summDuration := int32(0)
	formatedResult := make([]map[string]interface{}, 0)
	if fileSize > fileSizeLimit {
		formatedResult, summDuration = t.processLargeFile(filePath, formatedResult, summDuration)
	} else {
		formatedResult, summDuration = t.processSmallFile(filePath, formatedResult, summDuration)
	}

	t.logger.Infof("Whisper transcription end summDuration is : %v", summDuration)

	//remove mp3 file
	err = os.Remove(filePath)
	if err != nil {
		t.logger.Errorf("Failed to remove audio file: %v", err)
	}

	err = mrequests.UpdateCompanySttStatsByRoom(t.room, summDuration)
	if err != nil {
		t.logger.Errorf("Failed to update company stt stats: %v", err)
	}
	err = mrequests.UpdateTranscribeText(t.room, formatedResult)
	if err != nil {
		t.logger.Errorf("Failed to update transcribe text: %v", err)
	}
	mrequests.UpdateTranscribeTextStatus(t.room, "success")
	t.logger.Info("Transcription Results:", formatedResult)
}

func (t *OpenAiTranscribe) processLargeFile(filePath string, formatedResult []map[string]interface{}, summDuration int32) ([]map[string]interface{}, int32) {
	cfg := config.GetConfig()
	dirPathForCreation := cfg.TmpFilesPath + "/" + t.room
	err := utils.CreateDirByPath(dirPathForCreation)
	if err != nil {
		t.logger.Errorf("Failed to create directory: %v", err)
	}

	countOfSegments := t.DevideFileByTime(filePath, t.room, segmentSize)
	t.logger.Info("countOfSegments is : ", countOfSegments)

	for i := 0; i < countOfSegments; i++ {
		filePath = dirPathForCreation + "/segment_00" + fmt.Sprint(i) + ".mp3"
		tmpFormatedResult, duration, err := t.MakeWhisperFileTranscribe(filePath)
		if err != nil {
			t.logger.Infof("Failed to transcribe audio file: %v", err)
			continue
		}
		formatedResult = append(formatedResult, tmpFormatedResult...)
		summDuration += duration
		t.logger.Infof("Transcription for file proceed: %+v\n", filePath)
	}
	err = utils.RemoveDirWithFiles(dirPathForCreation)
	if err != nil {
		t.logger.Errorf("Failed to remove directory: %v", err)
	}
	t.logger.Infof("Transcription Results: %+v\n", formatedResult)

	return formatedResult, summDuration
}

func (t *OpenAiTranscribe) processSmallFile(filePath string, formatedResult []map[string]interface{}, summDuration int32) ([]map[string]interface{}, int32) {
	formatedResult, summDuration, err := t.MakeWhisperFileTranscribe(filePath)
	if err != nil {
		t.logger.Errorf("Failed to transcribe audio file: %v", err)
		mrequests.UpdateTranscribeTextStatus(t.room, "error")
		return formatedResult, summDuration
	}
	return formatedResult, summDuration
}

func (t *OpenAiTranscribe) FormatTranscribResult(resp *openai.AudioResponse) []map[string]interface{} {
	transcribeResult := make([]map[string]interface{}, 0)
	for _, segment := range resp.Segments {
		msgId := uuid.New().String()
		itemResult := map[string]interface{}{
			"id":    msgId,
			"msgID": msgId,
			"lang":  t.lang,
			"msg":   segment.Text,
		}
		transcribeResult = append(transcribeResult, itemResult)
	}

	return transcribeResult
}

func (t *OpenAiTranscribe) DownloadAudioFile(fileUrl string) (string, error) {
	cfg := config.GetConfig()
	// Send a GET request to the file URL
	resp, err := http.Get(fileUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Create a new file in the current directory
	fileName := uuid.New().String() + ".mp3"
	filePath := cfg.TmpFilesPath + "/" + fileName
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Write the response body to file
	t.logger.Info("resp.Body: ", resp.Body)
	_, err = io.Copy(file, resp.Body)

	//filePath := cfg.TmpFilesPath + "65c65ab594b860.75690686.mp3"

	return filePath, err
}
