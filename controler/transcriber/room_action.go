package transcriber

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
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
		c.AbortWithError(500, err)
		return
	}

	call, err := mrequests.GetCallByRoom(data.Room)
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	uids := call["uids"].([]string)
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
	err = mrequests.UpdateCallByBsonFilter(bson.M{
		"url": data.Room,
	}, bson.M{
		"$set": bson.M{
			"stt_for_all":           sttForAll,
			"stt_for_all_lang":      data.Lang.Code,
			"stt_for_all_lang_text": data.Lang.Text,
		}})

	if err != nil {
		logger.Println("Error when try update call stt_for_all by url: ", err)
	}
	// --------------------------------------------

	// get uids from call and switch on action (start/stop) for each user
	for i := range uids {
		uid := uids[i]
		logger.Printf("Start transcriber for uid: %v", uid)
		if publishAction == "sttForAllStart" {
			_, err := StartTranscriberAction(&TranscriberData{
				Room: data.Room,
				Lang: data.Lang.Code,
				Uid:  uid,
			})
			if err != nil {
				logger.Errorf("Error in StartTranscriberAction: %v for uid: %v", err, uid)
			}
		} else if publishAction == "sttForAllStop" {
			_, err := StopTranscriberAction(data.Room, uid)
			if err != nil {
				logger.Errorf("Error in StopTranscriberAction: %v for uid: %v", err, uid)
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
