package controler

import (
	"encoding/json"
	"fmt"
	"lkrouter/config"
	"lkrouter/utils"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type RecordEndedData struct {
	Room      string `json:"room"`
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
		c.AbortWithError(400, err)
		return
	}
	data.Timestamp = fmt.Sprintf("%d", time.Now().Unix())
	data.HashCode = utils.EncryptAuthData(cfg.WebhookUsername, cfg.WebhookPassword, data.Timestamp)
	data.Event = "recordUrl"

	log.Printf("We send data to the next server: ", data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		c.AbortWithError(400, err)
		return
	}

	log.Printf("Get data about ended record: %s", data)
	err = utils.SendWebhookData(jsonData, cfg.WebhookURL, cfg.WebhookUsername, cfg.WebhookPassword)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, data)
}
