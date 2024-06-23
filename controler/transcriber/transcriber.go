package transcriber

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"lkrouter/communications"
	"lkrouter/domain"
	"lkrouter/pkg/awslogs"
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

		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberStartController",
			"message": "Error binding json to TranscriberData",
			"type":    awslogs.MsgTypeError,
		})

		c.AbortWithError(500, err)
		return
	}

	transcResponse, err := StartTranscriberAction(&data)
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberStartController",
			"message": fmt.Sprintf("Error in StartTranscriberAction: %v, with data: %+v", err, data),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})

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

		awslogs.AddSLog(map[string]string{
			"func":    "StartTranscriberAction",
			"message": fmt.Sprintf("Livekit error in GetAudioTrackID: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})

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
	// set lang to sid_langs field in call collection
	err := mrequests.UpdateCallByBsonFilter(
		bson.M{"url": data.Room},
		bson.M{"$set": bson.M{
			"stt_user_lang." + data.Uid:   data.Lang,
			"stt_user_active." + data.Uid: true,
		}})
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "saveUserLang",
			"message": fmt.Sprintf("Error in saveUserLang: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	// send livekit user Metadata
	lkService := livekitserv.NewLiveKitService()
	_, err = lkService.UpdateUserMData(data.Room, data.Uid, map[string]interface{}{
		"sttActive": true,
		"sttLang":   data.Lang,
	})
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "saveUserLang",
			"message": fmt.Sprintf("Error in UpdateUserMData: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	// send message to client bs to start transcribe
	err = SendTranscriberStartMessage(data.Room, data.Uid, data.Lang)
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "saveUserLang",
			"message": fmt.Sprintf("Error in SendMessageToParticipant: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}
}

func SendTranscriberStartMessage(room string, uid string, lang string) error {
	sttMsg := domain.LKSttMsg{
		Type: domain.RoomActionTopic,
		Payload: domain.SttActionMsg{
			Action: "sttStart",
			Lang:   lang,
		},
	}
	sttMsgBytes, _ := json.Marshal(sttMsg)
	return livekitserv.NewLiveKitService().SendMessageToParticipant(room, uid, sttMsgBytes, domain.RoomActionTopic)
}

func TranscriberStopController(c *gin.Context) {
	data := TranscriberData{}
	if err := c.BindJSON(&data); err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberStopController",
			"message": "Error binding json to TranscriberData",
			"type":    awslogs.MsgTypeError,
		})

		c.AbortWithError(500, err)
		return
	}

	// get url from transcriber service
	transcResponse, err := StopTranscriberAction(data.Room, data.Uid)
	if err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberStopController",
			"message": fmt.Sprintf("Error in StopTranscriberAction: %v, with data: %+v", err, data),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})

		c.AbortWithError(500, err)
		return
	}

	// send livekit user Metadata
	_, err = livekitserv.NewLiveKitService().UpdateUserMData(data.Room, data.Uid, map[string]interface{}{
		"sttActive": false,
	})
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberStopController",
			"message": fmt.Sprintf("Error in UpdateUserMData: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	err = SendTranscriberStopMessage(data.Room, data.Uid)
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "TranscriberStopController",
			"message": fmt.Sprintf("Error in SendTranscriberStopMessage: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	err = mrequests.UpdateCallByBsonFilter(
		bson.M{"url": data.Room},
		bson.M{"$set": bson.M{
			"stt_user_active." + data.Uid: false,
		}})
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "saveUserLang",
			"message": fmt.Sprintf("Error in UpdateCallByBsonFilter: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	c.JSON(200, transcResponse)
}

func SendTranscriberStopMessage(room string, uid string) error {
	sttMsg := domain.LKSttMsg{
		Type: domain.RoomActionTopic,
		Payload: domain.SttActionMsg{
			Action: "sttStop",
		},
	}
	sttMsgBytes, _ := json.Marshal(sttMsg)
	return livekitserv.NewLiveKitService().SendMessageToParticipant(room, uid, sttMsgBytes, domain.RoomActionTopic)
}

func StopTranscriberAction(room string, uid string) (*communications.TranscribeReq, error) {
	transcRequest := communications.TranscribeReq{
		Room:          room,
		ParticipantId: uid,
	}

	return transcRequest.TranscriberAction("stop")
}
