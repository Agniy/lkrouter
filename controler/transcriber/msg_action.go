package transcriber

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"lkrouter/communications"
	"lkrouter/pkg/mongodb/mrequests"
)

type RemoveMsgData struct {
	Room  string `json:"room"`
	MsgId string `json:"msgId"`
	Uid   string `json:"uid"`
}

func RemoveMsgController(c *gin.Context) {
	logger := logrus.New()
	data := RemoveMsgData{}
	if err := c.BindJSON(&data); err != nil {
		logAndAbort(c, "RemoveMsgController", "Error binding json to RoomActionData", err)
		return
	}

	transcRequest := communications.TranscribeReq{
		Room:          data.Room,
		MsgId:         data.MsgId,
		ParticipantId: data.Uid,
	}

	_, err := transcRequest.RemoveMsgAction()
	if err != nil {
		logAndAbort(c, "RemoveMsgController", "Error in RemoveMsgAction", err)
		return
	}

	logger.Info("RemoveSttMessage - msgId: ", data.MsgId)

	err = mrequests.UpdateCallByBsonFilter(bson.M{
		"url": data.Room,
	}, bson.M{"$pull": bson.M{"transcrib_text": bson.M{"msgID": data.MsgId}}})
	if err != nil {
		logger.Info("RemoveSttMessage: ", "Error when remove message from call by msgId: ", data.MsgId, err)
		logAndAbort(c, "RemoveMsgController", "Error when remove message from call by msgId", err)
	}

	c.JSON(200, gin.H{"message": "Message removed successfully"})
}
