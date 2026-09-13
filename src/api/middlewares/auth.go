package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mahdipeydai/taskmanager-go/api/helpers"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"github.com/mahdipeydai/taskmanager-go/pkg/service_errors"
	"github.com/mahdipeydai/taskmanager-go/services"
)

func Authentication(cfg *config.Config) gin.HandlerFunc {
	logger := logging.GetLogger(cfg)
	tokenService := services.GetTokenService(cfg, logger)

	return func(c *gin.Context) {
		var err error
		claimMap := map[string]interface{}{}
		header := c.GetHeader(constants.AuthorizationHeaderKey)
		if header == "" {
			err = &service_errors.ServiceError{EndUserMessage: service_errors.TokenRequired}
		} else {
			headerParts := strings.Split(header, " ")
			if len(headerParts) != 2 {
				err = &service_errors.ServiceError{EndUserMessage: service_errors.InvalidToken}
			} else {
				scheme := headerParts[0]
				token := headerParts[1]
				if scheme != constants.AuthorizationHeaderPrefix {
					err = &service_errors.ServiceError{EndUserMessage: service_errors.InvalidTokenSchema}
				} else {
					claimMap, err = tokenService.GetClaims(c.Request.Context(), token)
					if err != nil {
						switch err.(*jwt.ValidationError).Errors {
						case jwt.ValidationErrorExpired:
							err = &service_errors.ServiceError{EndUserMessage: service_errors.TokenExpired}
						default:
							err = &service_errors.ServiceError{EndUserMessage: service_errors.InvalidToken}
						}
					}
				}
			}
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.GenerateBaseResponseWithError(
				nil,
				false,
				-2,
				err,
			))
			return
		}

		c.Set(constants.UserIdKey, claimMap[constants.UserIdKey])
		c.Set(constants.UsernameKey, claimMap[constants.UsernameKey])
		c.Set(constants.FirstNameKey, claimMap[constants.FirstNameKey])
		c.Set(constants.LastNameKey, claimMap[constants.LastNameKey])
		c.Set(constants.EmailKey, claimMap[constants.EmailKey])
		c.Set(constants.MobileNumberKey, claimMap[constants.MobileNumberKey])
		c.Set(constants.RolesKey, claimMap[constants.RolesKey])
		c.Set(constants.ExpireTimeKey, claimMap[constants.ExpireTimeKey])

		c.Next()
	}
}

func Authorization(validRoles []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(constants.UserIdKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.GenerateBaseResponseWithError(
				nil,
				false,
				-300,
				service_errors.ServiceError{EndUserMessage: service_errors.PermissionDenied},
			))
			return
		}

		userRolesVal, exists := c.Get(constants.RolesKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.GenerateBaseResponseWithError(
				nil,
				false,
				-300,
				service_errors.ServiceError{EndUserMessage: service_errors.PermissionDenied},
			))
			return
		}
		if userRolesVal == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.GenerateBaseResponseWithError(
				nil,
				false,
				-300,
				service_errors.ServiceError{EndUserMessage: service_errors.PermissionDenied},
			))
			return
		}
		userRoles := userRolesVal.([]interface{})
		userRolesMap := map[string]interface{}{}
		for _, role := range userRoles {
			userRolesMap[role.(string)] = nil
		}

		for _, role := range validRoles {
			if _, ok := userRolesMap[role]; ok {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, helpers.GenerateBaseResponseWithError(
			nil,
			false,
			-300,
			service_errors.ServiceError{EndUserMessage: service_errors.PermissionDenied},
		))
		return
	}
}
