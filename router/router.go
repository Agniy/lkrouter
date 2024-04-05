package router

import (
	"lkrouter/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func init() {
	r = gin.Default()

	RouterConfig(r)
	EgressRouter(r)
	WebhookRouter(r)
}

func RouterConfig(r *gin.Engine) {

	cfg := config.GetConfig()

	expectedHost := cfg.Domain
	if cfg.Port != "80" {
		expectedHost += ":" + cfg.Port
	}

	if cfg.GinMode != "debug" {
		r.Use(func(c *gin.Context) {
			if c.Request.Host != expectedHost {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid host header"})
				return
			}
			c.Next()
		})
	}

	r.Use(gin.Recovery())
}

func GetRouter() *gin.Engine {
	return r
}
