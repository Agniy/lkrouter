package transcriber

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"lkrouter/communications"
	"lkrouter/pkg/livekitserv"
	"lkrouter/pkg/mongodb/mrequests"
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

	transcResponse, err := StartTranscriberAction(&data)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	// set lang to sid_langs field in call collection
	saveUserLang(&data)

	c.JSON(200, transcResponse)
}

func StartTranscriberAction(data *TranscriberData) (*communications.TranscribeReq, error) {
	trackId, err := livekitserv.NewLiveKitService().GetAudioTrackID(data.Room, data.Uid)
	if err != nil {
		return nil, err
	}

	// get url from transcriber service
	transcRequest := communications.TranscribeReq{
		Room:          data.Room,
		Lang:          data.Lang,
		TrackId:       trackId,
		ParticipantId: data.Uid,
	}

	return transcRequest.TranscriberAction("start")
}

func saveUserLang(data *TranscriberData) {
	logger := logrus.New()
	// set lang to sid_langs field in call collection
	err := mrequests.UpdateCallByBsonFilter(
		bson.M{"url": data.Room},
		bson.M{"$set": bson.M{
			"stt_user_lang." + data.Uid: data.Lang,
		}})
	if err != nil {
		logger.Errorf("Error in saveUserLang: %v", err)
	}
}

func TranscriberStopController(c *gin.Context) {
	data := TranscriberData{}
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(500, err)
		return
	}

	// get url from transcriber service
	transcResponse, err := StopTranscriberAction(data.Room, data.Uid)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	c.JSON(200, transcResponse)
}

func StopTranscriberAction(room string, uid string) (*communications.TranscribeReq, error) {
	transcRequest := communications.TranscribeReq{
		Room:          room,
		ParticipantId: uid,
	}

	return transcRequest.TranscriberAction("stop")
}
