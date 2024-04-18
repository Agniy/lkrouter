package router

import (
	"github.com/gin-gonic/gin"
)
import "lkrouter/controler"

func EgressRouter(r *gin.Engine) {
	egreessRouter := r.Group("/record")
	{
		egreessRouter.POST("/start/", func(c *gin.Context) {
			controler.EgressController(c)
		})
	}
}
