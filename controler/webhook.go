package controler

import "github.com/gin-gonic/gin"

type RecordEndedData struct {
	Room     string `json:"room"`
	AudioUrl string `json:"audioUrl"`
	VideoUrl string `json:"videoUrl"`
}

func RecordEndedController(c *gin.Context) {
	data := RecordEndedData{}
	if err := c.BindJSON(&data); err != nil {
		c.AbortWithError(400, err)
		return
	}

	c.JSON(200, data)
}
