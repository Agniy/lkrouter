package controler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lkrouter/pkg/livekitserv"
	rservice "lkrouter/service"
	"net/http"
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
	if !rservice.NewAuthService().CheckRoomPermission(c, data.Room) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err := livekitserv.NewLiveKitService().DeleteRoom(data.Room)
	if err != nil {
		fmt.Printf("Error stop room %v, error: %v \n", data.Room, err)
	}
	fmt.Printf("Room %v stopped \n", data.Room)

	response := CallStopResponse{
		Room:   data.Room,
		Status: "stopped",
	}
	c.JSON(http.StatusAccepted, &response)
}
