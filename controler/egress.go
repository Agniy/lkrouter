package controler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type EgressStartData struct {
	Room string `json:"room"`
}

type EgressStartResponse struct {
	Room   string `json:"room"`
	Status string `json:"status"`
}

func EgressController(c *gin.Context) {
	data := EgressStartData{}
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	fmt.Println(data)

	response := EgressStartResponse{
		Room:   data.Room,
		Status: "ok",
	}
	c.JSON(http.StatusAccepted, &response)
}
