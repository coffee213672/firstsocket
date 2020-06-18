package router

import (
	"go-chat/endpoint"

	"github.com/gin-gonic/gin"
)

// InitRouter ...
func InitRouter() *gin.Engine {
	r := gin.New()
	r.LoadHTMLFiles("index.html")
	r.GET("/room/:roomId", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})

	r.GET("/ws/:roomId", func(c *gin.Context) {
		roomID := c.Param("roomId")
		endpoint.ServeWs(c.Writer, c.Request, roomID)
	})
	return r
}
