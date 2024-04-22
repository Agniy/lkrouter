package controler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lkrouter/communications"
	"lkrouter/config"
	"lkrouter/pkg/egresserv"
	"strconv"
)

type TranscriberStartData struct {
	Room     string `json:"room"`
	Lang     string `json:"lang"`
	EgressId string `json:"egressId"`
}

func TranscriberStartController(c *gin.Context) {
	cfg := config.GetConfig()
	data := TranscriberStartData{}
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(400, err)
		return
	}

	// get url from transcriber service
	transcRequest := communications.TranscribeReq{
		Room: data.Room,
		Lang: data.Lang,
	}
	transcribePort, err := communications.GetTranscribePort(cfg.TranscribeAddr, &transcRequest)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	fmt.Println("Transcribe port: ", transcribePort, " for room: ", data.Room, " lang: ", data.Lang)

	wsUrl := "wss://127.0.0.1:" + strconv.Itoa(transcribePort)
	// start tracking egress
	egresInfo, err := egresserv.TrackEgressRequest(data.Room, data.Lang, wsUrl)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}
	data.EgressId = egresInfo.EgressId

	fmt.Println("Egress info: ", egresInfo)

	c.JSON(200, data)
}
