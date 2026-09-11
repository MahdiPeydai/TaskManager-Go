package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/golang-clean-web-api/api/helpers"
)

func ErrorHandler(c *gin.Context, err any) {
	if err, ok := err.(error); ok {
		httpResponse := helpers.GenerateBaseResponseWithError(nil, false, helpers.CustomRecovery, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, httpResponse)
		return
	}
	httpResponse := helpers.GenerateBaseResponseWithAnyError(nil, false, helpers.CustomRecovery, err)
	c.AbortWithStatusJSON(http.StatusInternalServerError, httpResponse)
}
