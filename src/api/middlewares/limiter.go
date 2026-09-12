package middlewares

import (
	"net/http"

	"github.com/didip/tollbooth/v8"
	"github.com/didip/tollbooth/v8/limiter"
	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/helpers"
)

func LimitByRequest(rateLimit float64) gin.HandlerFunc {
	lmt := tollbooth.NewLimiter(rateLimit, nil)
	lmt.SetIPLookup(limiter.IPLookup{
		Name:           "X-Real-IP",
		IndexFromRight: 0,
	})
	return func(c *gin.Context) {
		if c.GetHeader("X-Real-IP") == "" {
			c.Request.Header.Set("X-Real-IP", c.ClientIP())
		}

		err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusTooManyRequests,
				helpers.GenerateBaseResponseWithError(
					nil,
					false,
					-100,
					err,
				),
			)
			return
		}
		c.Next()
	}
}
