package router

import "github.com/gin-gonic/gin"
import "lkrouter/controler"

func TranscribeRouter(r *gin.Engine) {
	r.POST("/transcribefile", func(c *gin.Context) {
		controler.WhisperTranscribeFileController(c)
	})
}
