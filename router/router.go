package router

import (
	"github.com/gin-gonic/gin"
	"lkrouter/config"
)

var r *gin.Engine

func init() {
	r = gin.Default()

	RouterConfig(r)
	EgressRouter(r)
	WebhookRouter(r)
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func RouterConfig(r *gin.Engine) {

	cfg := config.GetConfig()

	expectedHost := cfg.Domain
	if cfg.Port != "80" {
		expectedHost += ":" + cfg.Port
	}

	//if cfg.GinMode != "debug" {
	//	r.Use(func(c *gin.Context) {
	//		if c.Request.Host != expectedHost {
	//			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid host header"})
	//			return
	//		}
	//		c.Next()
	//	})
	//}

	r.Use(CORSMiddleware())
	r.Use(gin.Recovery())
}

func GetRouter() *gin.Engine {
	return r
}
