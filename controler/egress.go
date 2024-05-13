package controler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lkrouter/pkg/egresserv"
	"lkrouter/pkg/livekitserv"
	"net/http"
)

type EgressStartData struct {
	Room    string `json:"room"`
	Company string `json:"company"`
}

type EgressStartResponse struct {
	Room     string `json:"room"`
	Status   string `json:"status"`
	EgressID string `json:"egressID"`
}

func EgressController(c *gin.Context) {
	data := EgressStartData{}
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	eggressID, err := egresserv.StartTrackEgress(data.Room, data.Company)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("Try to send to room: ", data.Room, " new metadata: \"rec\": \"true\"")

	room, err := livekitserv.NewLiveKitService().UpdateRoomMData(data.Room, map[string]interface{}{
		"rec": true,
	})

	if err != nil {
		fmt.Println("Error updating room metadata", err)
	}

	fmt.Println("Room metadata updated", room.Metadata)

	response := EgressStartResponse{
		Room:     data.Room,
		Status:   "ok",
		EgressID: eggressID,
	}
	c.JSON(http.StatusAccepted, &response)
}
