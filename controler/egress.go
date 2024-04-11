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
	eggressID := egresserv.StartTrackEgress(data.Room, data.Company)

	_, err := livekitserv.NewLiveKitService().UpdateRoomMData(data.Room, map[string]string{
		"rec": "true",
	})

	if err != nil {
		fmt.Println("Error updating room metadata", err)
	}

	response := EgressStartResponse{
		Room:     data.Room,
		Status:   "ok",
		EgressID: eggressID,
	}
	c.JSON(http.StatusAccepted, &response)
}
