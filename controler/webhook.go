package controler

import (
	"encoding/json"
	"fmt"
	"lkrouter/config"
	"lkrouter/pkg/awslogs"
	"lkrouter/pkg/livekitserv"
	"lkrouter/pkg/redisdb"
	"lkrouter/utils"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type RecordEndedData struct {
	Room      string `json:"room"`
	Company   string `json:"company"`
	AudioUrl  string `json:"audioUrl"`
	VideoUrl  string `json:"videoUrl"`
	Timestamp string `json:"timestamp"`
	HashCode  string `json:"hashCode"`
	Event     string `json:"event"`
	FileSize  int64  `json:"fileSize"`
}

func RecordEndedController(c *gin.Context) {
	cfg := config.GetConfig()
	data := RecordEndedData{}
	if err := c.BindJSON(&data); err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "RecordEndedController",
			"message": fmt.Sprintf("Error binding json to RecordEndedData: %v", err),
			"type":    awslogs.MsgTypeError,
		})

		c.AbortWithError(400, err)
		return
	}

	//set record status to "stopped"
	err := redisdb.SetRoomRecordStatus(data.Room, "stopped", 10*time.Minute)
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "RecordEndedController",
			"message": fmt.Sprintf("Redis Error saving record status: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	_, err = livekitserv.NewLiveKitService().UpdateRoomMData(data.Room, map[string]interface{}{
		"rec":        false,
		"rec-status": "stopped",
	})
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "RecordEndedController",
			"message": fmt.Sprintf("Error updating room metadata: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	data.Timestamp = fmt.Sprintf("%d", time.Now().Unix())
	data.HashCode = utils.EncryptAuthData(cfg.WebhookUsername, cfg.WebhookPassword, data.Timestamp)
	data.Event = "recordUrl"

	log.Printf("We send data to the next server: ", data)

	jsonData, err := json.Marshal(data)
	if err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "RecordEndedController",
			"message": fmt.Sprintf("Error marshalling data: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})

		c.AbortWithError(400, err)
		return
	}

	log.Printf("Get data about ended record: %s", data)
	err = utils.SendWebhookData(jsonData, cfg.WebhookURL, cfg.WebhookUsername, cfg.WebhookPassword)
	if err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "RecordEndedController",
			"message": fmt.Sprintf("Error sending webhook data: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})

		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, data)
}
