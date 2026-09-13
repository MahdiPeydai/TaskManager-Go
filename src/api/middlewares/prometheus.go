package middlewares

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/pkg/metrics"
)

func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown path"
		}
		method := c.Request.Method

		c.Next()
		status := c.Writer.Status()

		metrics.HttpDuration.WithLabelValues(method, path, strconv.Itoa(status)).
			Observe(float64(time.Since(start) / time.Millisecond))

		metrics.RequestsTotal.WithLabelValues(method, path, strconv.Itoa(status)).
			Inc()
	}
}
