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
		logAndAbort(c, "RoomActionController", "Error binding json to RoomActionData", err)
		return
	}

	publishAction, sttForAll := getPublishActionAndSttForAll(data.Action)

	err := updateCallFieldSttForAllAndSttForAllLang(data, sttForAll)
	if err != nil {
		awslogs.LogError("RoomActionController", fmt.Sprintf("MongoDB error in UpdateCallByBsonFilter: %v", err), data.Room)
	}

	_, err = livekitserv.NewLiveKitService().UpdateRoomMData(data.Room, map[string]interface{}{
		"sttForAll":     sttForAll,
		"sttForAllLang": data.Lang.Code,
	})
	if err != nil {
		awslogs.LogError("RoomActionController", fmt.Sprintf("Livekit error update sttForAll in UpdateRoomMData: %v", err), data.Room)
	}

	handleActionForUsers(data, publishAction, logger)

	res := RoomActionResponse{
		Status: "ok",
		Action: publishAction,
		Lang:   data.Lang.Code,
		Text:   data.Lang.Text,
	}
	c.JSON(200, res)
}

func logAndAbort(c *gin.Context, funcName string, message string, err error) {
	awslogs.AddSLog(map[string]string{
		"func":    funcName,
		"message": message,
		"type":    awslogs.MsgTypeError,
	})
	c.AbortWithError(500, err)
}

func getPublishActionAndSttForAll(action string) (string, bool) {
	publishAction := ""
	sttForAll := false
	if action == "start" {
		publishAction = "sttForAllStart"
		sttForAll = true
	} else if action == "stop" {
		publishAction = "sttForAllStop"
	}
	return publishAction, sttForAll
}

func updateCallFieldSttForAllAndSttForAllLang(data RoomActionData, sttForAll bool) error {
	return mrequests.UpdateCallByBsonFilter(bson.M{
		"url": data.Room,
	}, bson.M{
		"$set": bson.M{
			"stt_for_all":           sttForAll,
			"stt_for_all_lang":      data.Lang.Code,
			"stt_for_all_lang_text": data.Lang.Text,
		}})
}

func handleActionForUsers(data RoomActionData, publishAction string, logger *logrus.Logger) {
	lkServ := livekitserv.NewLiveKitService()
	userList, err := lkServ.RealParticipantsByRoom(data.Room)
	if err != nil {
		awslogs.LogError(
			"RoomActionController",
			fmt.Sprintf("Livekit error in RealParticipantsByRoom: %v", err),
			data.Room)
	}

	for i := range userList {
		uInfo := userList[i]
		uid := uInfo.Identity
		logger.Printf("Start transcriber for uid: %v", uid)
		if publishAction == "sttForAllStart" {
			handleStartAction(data, uid)
		} else if publishAction == "sttForAllStop" {
			handleStopAction(data, uid)
		}
	}
}

func handleStartAction(data RoomActionData, uid string) {
	_, err := StartTranscriberAction(&TranscriberData{
		Room: data.Room,
		Lang: data.Lang.Code,
		Uid:  uid,
	})
	if err != nil {
		awslogs.LogError(
			"RoomActionController",
			fmt.Sprintf("Error in StartTranscriberAction: %v, with data: %+v", err, data),
			data.Room)
	}

	err = SendTranscriberStartMessage(data.Room, uid, data.Lang.Code)
	if err != nil {
		awslogs.LogError(
			"RoomActionController",
			fmt.Sprintf("Error in SendTranscriberStartMessage: %v for uid: %v", err, uid),
			data.Room)
	}
}

func handleStopAction(data RoomActionData, uid string) {
	_, err := StopTranscriberAction(data.Room, uid)
	if err != nil {
		awslogs.LogError(
			"RoomActionController",
			fmt.Sprintf("Error in StopTranscriberAction: %v for uid: %v", err, uid),
			data.Room)
	}

	err = SendTranscriberStopMessage(data.Room, uid)
	if err != nil {
		awslogs.LogError(
			"RoomActionController",
			fmt.Sprintf("Error in SendTranscriberStopMessage: %v for uid: %v", err, uid),
			data.Room)
	}
}
