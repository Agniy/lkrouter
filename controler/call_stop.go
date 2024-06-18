package controler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"lkrouter/domain"
	"lkrouter/pkg/awslogs"
	"lkrouter/pkg/livekitserv"
	rservice "lkrouter/service"
	"net/http"
	"time"
)

type CallStopData struct {
	Room string `json:"room"`
}

type CallStopResponse struct {
	Room   string `json:"room"`
	Status string `json:"status"`
}

func CallStopController(c *gin.Context) {

	data := CallStopData{}
	if err := c.BindJSON(&data); err != nil {

		awslogs.AddSLog(map[string]string{
			"func":    "CallStopController",
			"message": fmt.Sprintf("Error binding json to CallStopData: %v", err),
			"type":    awslogs.MsgTypeError,
		})

		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// check user permission
	hasPermission, err := rservice.NewAuthService().CheckRoomPermission(c, data.Room)
	if !hasPermission {
		msg := fmt.Sprintf("User has no permission to stop room %v, error: %v", data.Room, err)
		fmt.Println(msg)

		awslogs.AddSLog(map[string]string{
			"func":    "CallStopController",
			"message": fmt.Sprintf(msg),
			"room":    data.Room,
			"type":    awslogs.MsgTypeWarn,
		})

		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Send message to all participants in the room
	msg, _ := json.Marshal(domain.RoomActionMessage{Action: "roomStop"})
	err = livekitserv.NewLiveKitService().SendMessageToParticipants(data.Room, msg, "room_action")
	if err != nil {
		awslogs.AddSLog(map[string]string{
			"func":    "CallStopController",
			"message": fmt.Sprintf("Livekit error sending message to participants in room %v: %v", data.Room, err),
			"room":    data.Room,
			"type":    awslogs.MsgTypeError,
		})
	}

	// Call DeleteRoom after a delay of 3 seconds
	time.AfterFunc(3*time.Second, func() {
		err = livekitserv.NewLiveKitService().DeleteRoom(data.Room)
		if err != nil {
			msg := fmt.Sprintf("Livekit error stop room %v, error: %v", data.Room, err)

			awslogs.AddSLog(map[string]string{
				"func":    "CallStopController",
				"message": msg,
				"room":    data.Room,
				"type":    awslogs.MsgTypeError,
			})

		} else {
			awslogs.AddSLog(map[string]string{
				"func":    "CallStopController",
				"message": fmt.Sprintf("Room %v stopped", data.Room),
				"room":    data.Room,
				"type":    awslogs.MsgTypeInfo,
			})

		}
	})

	response := CallStopResponse{
		Room:   data.Room,
		Status: "stopped",
	}
	c.JSON(http.StatusAccepted, &response)
}
