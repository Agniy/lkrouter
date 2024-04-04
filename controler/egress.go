package controler

import (
	"github.com/gin-gonic/gin"
	"lkrouter/pkg/egresserv"
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

	response := EgressStartResponse{
		Room:     data.Room,
		Status:   "ok",
		EgressID: eggressID,
	}
	c.JSON(http.StatusAccepted, &response)
}
