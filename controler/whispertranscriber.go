package controler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lkrouter/pkg/awslogs"
	"lkrouter/pkg/mongodb/mrequests"
	"lkrouter/pkg/transcribe"
	"net/http"
)

type TranscribeFileData struct {
	Room   string `json:"room"`
	Lang   string `json:"lang"`
	Prompt string `json:"prompt"`
}

const (
	STATUS_PROGRESS string = "progress"
	STATUS_SUCCESS  string = "success"
	STATUS_ERROR    string = "error"
)

func WhisperTranscribeFileController(c *gin.Context) {
	response := make(map[string]string)

	transcribeData := TranscribeFileData{}
	if err := c.BindJSON(&transcribeData); err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "WhisperTranscribeFileController",
			"message": "Error binding json to TranscribeFileData",
			"type":    awslogs.MsgTypeError,
		})

		c.AbortWithError(400, err)
		return
	}

	// get call by room
	call, err := mrequests.GetCallByRoom(transcribeData.Room)
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "WhisperTranscribeFileController",
			"message": fmt.Sprintf("Mogodb error when try to get room %v", transcribeData.Room),
			"type":    awslogs.MsgTypeError,
			"room":    transcribeData.Room,
		})
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	// check company sttLimit
	if call["companyId"] == nil {
		awslogs.AddSLog(map[string]string{
			"func":    "WhisperTranscribeFileController",
			"message": fmt.Sprintf("Error when try to get company id by room %v, companyId is nil", transcribeData.Room),
			"type":    awslogs.MsgTypeError,
			"room":    transcribeData.Room,
		})
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	// check company sttLimit in seconds
	var callStt int32 = 0
	if call["audioDuration"] != nil {
		callStt = int32(call["audioDuration"].(float64))
	}

	// TODO add check for whisper limit
	err = mrequests.CheckCompanySttLimit(call["companyId"].(string), callStt)
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "WhisperTranscribeFileController",
			"message": fmt.Sprintf("SttLimit of company sttCurrent + %v limit out of border by room %v \n", callStt, transcribeData.Room),
			"type":    awslogs.MsgTypeWarn,
			"room":    transcribeData.Room,
		})
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	if call["audioUrl"] == nil {
		awslogs.AddSLog(map[string]string{
			"func":    "WhisperTranscribeFileController",
			"message": fmt.Sprintf("Error when try to get audioUrl from room %v, audioUrl is nil", transcribeData.Room),
			"type":    awslogs.MsgTypeError,
			"room":    transcribeData.Room,
		})
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	if call["file_transcribe_status"] == STATUS_PROGRESS {
		response["status"] = STATUS_PROGRESS
		response["message"] = "Transcribe is in progress"
		c.JSON(http.StatusOK, response)
		return
	} else {
		transcribe.SendWorkTask(map[string]interface{}{
			"room":   transcribeData.Room,
			"prompt": transcribeData.Prompt,
			"type":   "whisper",
		})
		response["status"] = STATUS_PROGRESS
		response["message"] = "Transcribe is in progress"
		c.JSON(http.StatusOK, response)
		return
	}
}
