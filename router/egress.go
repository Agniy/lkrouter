package router

import "github.com/gin-gonic/gin"
import "lkrouter/controler"

func EgressRouter(r *gin.Engine) {
	egreessRouter := r.Group("/record")
	{
		egreessRouter.OPTIONS("/start/", func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
			c.Writer.Header().Set("Content-Length", "0")
			c.Writer.WriteHeader(204)
		})

		egreessRouter.POST("/start/", func(c *gin.Context) {
			controler.EgressController(c)
		})
	}
}
