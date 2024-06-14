package router

import (
	"github.com/gin-gonic/gin"
	"lkrouter/controler/transcriber"
)

func TranscriberRouter(r *gin.Engine) {
	egreessRouter := r.Group("/transcriber")
	{
		egreessRouter.POST("/start", func(c *gin.Context) {
			transcriber.TranscriberStartController(c)
		})

		egreessRouter.POST("/stop", func(c *gin.Context) {
			transcriber.TranscriberStopController(c)
		})

		egreessRouter.POST("/room_action", func(c *gin.Context) {
			transcriber.RoomActionController(c)
		})

	}
}
