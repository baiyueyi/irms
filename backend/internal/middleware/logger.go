package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		rid, _ := c.Get("request_id")
		log.Printf("rid=%v method=%s path=%s status=%d latency=%s", rid, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
	}
}

