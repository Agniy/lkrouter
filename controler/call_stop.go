package controler

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"lkrouter/domain"
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
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// check user permission
	hasPermission, err := rservice.NewAuthService().CheckRoomPermission(c, data.Room)
	if !hasPermission {
		fmt.Printf("User has no permission to stop room %v, error: %v \n", data.Room, err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Send message to all participants in the room
	msg, _ := json.Marshal(domain.RoomActionMessage{Action: "roomStop"})
	err = livekitserv.NewLiveKitService().SendMessageToParticipants(
		data.Room, msg, "room_action")

	// Call DeleteRoom after a delay of 3 seconds
	time.AfterFunc(3*time.Second, func() {
		err = livekitserv.NewLiveKitService().DeleteRoom(data.Room)
		if err != nil {
			fmt.Printf("Error stop room %v, error: %v \n", data.Room, err)
		}
		fmt.Printf("Room %v stopped \n", data.Room)
	})

	response := CallStopResponse{
		Room:   data.Room,
		Status: "stopped",
	}
	c.JSON(http.StatusAccepted, &response)
}
