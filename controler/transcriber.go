package controler

import (
	"github.com/gin-gonic/gin"
	"lkrouter/communications"
	"lkrouter/pkg/livekitserv"
)

type TranscriberData struct {
	Room string `json:"room"`
	Lang string `json:"lang"`
	Uid  string `json:"uid"`
}

func TranscriberStartController(c *gin.Context) {
	data := TranscriberData{}
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(500, err)
		return
	}

	trackId, err := livekitserv.NewLiveKitService().GetAudioTrackID(data.Room, data.Uid)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// get url from transcriber service
	transcRequest := communications.TranscribeReq{
		Room:          data.Room,
		Lang:          data.Lang,
		TrackId:       trackId,
		ParticipantId: data.Uid,
	}

	transcResponse, err := transcRequest.TranscriberAction("start")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, transcResponse)
}

func TranscriberStopController(c *gin.Context) {
	data := TranscriberData{}
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(500, err)
		return
	}

	// get url from transcriber service
	transcRequest := communications.TranscribeReq{
		Room:          data.Room,
		ParticipantId: data.Uid,
	}

	transcResponse, err := transcRequest.TranscriberAction("stop")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, transcResponse)
}
