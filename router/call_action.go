package router

import "github.com/gin-gonic/gin"
import "lkrouter/controler"

func CallStopRouter(r *gin.Engine) {
	egreessRouter := r.Group("/call")
	{
		egreessRouter.POST("/stop", func(c *gin.Context) {
			controler.CallStopController(c)
		})
	}
}
