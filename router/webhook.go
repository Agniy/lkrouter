package router

import "github.com/gin-gonic/gin"
import "lkrouter/controler"

func WebhookRouter(r *gin.Engine) {
	egreessRouter := r.Group("/webhook")
	{
		egreessRouter.POST("/record/end/", func(c *gin.Context) {
			controler.RecordEndedController(c)
		})
	}
}
