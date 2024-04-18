package controler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lkrouter/pkg/livekitserv"
	"net/http"
)

func CallsListController(c *gin.Context) {
	rooms, err := livekitserv.NewLiveKitService().GetAllActiveCalls()
	if err != nil {
		fmt.Println("Error getting all rooms", err)
	}
	fmt.Println("Room: ", rooms)

	response := struct{}{}
	c.JSON(http.StatusAccepted, &response)
}
