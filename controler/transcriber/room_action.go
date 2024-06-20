package transcriber

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"lkrouter/pkg/awslogs"
	"lkrouter/pkg/livekitserv"
	"lkrouter/pkg/mongodb/mrequests"
)

type LangCodeText struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

type RoomActionData struct {
	Room   string       `json:"room"`
	Lang   LangCodeText `json:"lang"`
	Action string       `json:"action"`
	Uid    string       `json:"uid"`
}

type RoomActionResponse struct {
	Status string `json:"status"`
	Action string `json:"action"`
	Lang   string `json:"lang"`
	Text   string `json:"text"`
}

func RoomActionController(c *gin.Context) {
	logger := logrus.New()
	data := RoomActionData{}
	if err := c.BindJSON(&data); err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "RoomActionController",
			"message": "Error binding json to RoomActionData",
			"type":    awslogs.MsgTypeError,
		})

		c.AbortWithError(500, err)
		return
	}

	publishAction := ""
	sttForAll := false
	if data.Action == "start" {
		publishAction = "sttForAllStart"
		sttForAll = true
	} else if data.Action == "stop" {
		publishAction = "sttForAllStop"
	}

	//update call field stt_for_all and stt_for_all_lang
	// --------------------------------------------
	err := mrequests.UpdateCallByBsonFilter(bson.M{
		"url": data.Room,
	}, bson.M{
		"$set": bson.M{
			"stt_for_all":           sttForAll,
			"stt_for_all_lang":      data.Lang.Code,
			"stt_for_all_lang_text": data.Lang.Text,
		}})
	if err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "RoomActionController",
			"message": fmt.Sprintf("MongoDB error in UpdateCallByBsonFilter: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	// set livekit user lang
	_, err = livekitserv.NewLiveKitService().UpdateRoomMData(data.Room, map[string]interface{}{
		"sttForAll":     sttForAll,
		"sttForAllLang": data.Lang.Code,
	})

	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "RoomActionController",
			"message": fmt.Sprintf("Livekit error update sttForAll in UpdateRoomMData: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}
	// --------------------------------------------

	lkServ := livekitserv.NewLiveKitService()
	userList, err := lkServ.RealParticipantsByRoom(data.Room)
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "RoomActionController",
			"message": fmt.Sprintf("Livekit error in RealParticipantsByRoom: %v", err),
			"type":    awslogs.MsgTypeError,
			"room":    data.Room,
		})
	}

	// get uids from call and switch on action (start/stop) for each user
	for i := range userList {
		uInfo := userList[i]
		uid := uInfo.Identity
		logger.Printf("Start transcriber for uid: %v", uid)
		if publishAction == "sttForAllStart" {
			_, err = StartTranscriberAction(&TranscriberData{
				Room: data.Room,
				Lang: data.Lang.Code,
				Uid:  uid,
			})
			if err != nil {
				awslogs.AddSLog(map[string]string{
					"func":    "RoomActionController",
					"message": fmt.Sprintf("Error in StartTranscriberAction: %v, with data: %+v", err, data),
					"type":    awslogs.MsgTypeError,
					"room":    data.Room,
				})
			}
		} else if publishAction == "sttForAllStop" {
			_, err = StopTranscriberAction(data.Room, uid)
			if err != nil {
				awslogs.AddSLog(map[string]string{
					"func":    "RoomActionController",
					"message": fmt.Sprintf("Error in StopTranscriberAction: %v for uid: %v", err, uid),
					"type":    awslogs.MsgTypeError,
					"room":    data.Room,
				})
			}
		}
	}

	res := RoomActionResponse{
		Status: "ok",
		Action: publishAction,
		Lang:   data.Lang.Code,
		Text:   data.Lang.Text,
	}
	c.JSON(200, res)

}
