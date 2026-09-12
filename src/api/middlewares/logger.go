package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func LogRequestResponse(logger logging.LoggerInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.FullPath(), "swagger") {
			c.Next()
		} else {
			blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			start := time.Now()
			path := c.Request.URL.Path
			raw := c.Request.URL.RawQuery
			if raw != "" {
				path += "?" + raw
			}

			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				extras := map[logging.ExtraKey]interface{}{logging.ErrorMessage: err.Error()}
				logger.Error(logging.Internal, logging.API, "Failed to read request body.", extras)
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			err = c.Request.Body.Close()
			if err != nil {
				extras := map[logging.ExtraKey]interface{}{logging.ErrorMessage: err.Error()}
				logger.Error(logging.Internal, logging.API, "Failed to close request body reader.", extras)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			c.Writer = blw
			c.Next()

			end := time.Now()
			latency := end.Sub(start)

			extras := map[logging.ExtraKey]interface{}{}
			extras[logging.Path] = path
			extras[logging.ClientIp] = c.ClientIP()
			extras[logging.Method] = c.Request.Method
			extras[logging.StatusCode] = c.Writer.Status()
			extras[logging.ErrorMessage] = c.Errors.ByType(gin.ErrorTypeAny).String()
			extras[logging.ResponseBody] = string(bodyBytes)
			extras[logging.ResponseBody] = blw.body.String()
			extras[logging.BodySize] = c.Writer.Size()
			extras[logging.Latency] = latency

			logger.Info(logging.RequestResponse, logging.API, "", extras)
		}
	}
}
