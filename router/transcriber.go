package router

import "github.com/gin-gonic/gin"
import "lkrouter/controler"

func TranscriberRouter(r *gin.Engine) {
	egreessRouter := r.Group("/transcriber")
	{
		egreessRouter.POST("/start", func(c *gin.Context) {
			controler.TranscriberStartController(c)
		})

		egreessRouter.POST("/stop", func(c *gin.Context) {
			controler.TranscriberStopController(c)
		})
	}
}
