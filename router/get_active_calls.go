package router

import "github.com/gin-gonic/gin"
import "lkrouter/controler"

func CallsRouter(r *gin.Engine) {
	egreessRouter := r.Group("/calls")
	{
		egreessRouter.POST("/all/", func(c *gin.Context) {
			controler.CallsListController(c)
		})
	}
}
