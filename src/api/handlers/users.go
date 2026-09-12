package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/mahdipeydai/taskmanager-go/api/helpers"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/data/db"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"github.com/mahdipeydai/taskmanager-go/services"
)

type UsersHandler struct {
	cfg         *config.Config
	userService *services.UsersService
}

func GetUsersHandler(cfg *config.Config) *UsersHandler {
	logger := logging.GetLogger(cfg)
	return &UsersHandler{
		userService: services.GetUsersService(db.GetDB(), cfg, logger),
		cfg:         cfg,
	}
}

// RegisterByUsername godoc
//
//	@Summary		Register By Username
//	@Description	Register By Username
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			Request	body		dto.RegisterUserByUsernameRequest	true	"Register by username"
//	@Success		201		{object}	helpers.BaseHTTPResponse			"Success"
//	@Failure		400		{object}	helpers.BaseHTTPResponse			"Failed"
//	@Failure		404		{object}	helpers.BaseHTTPResponse			"Failed"
//	@Failure		409		{object}	helpers.BaseHTTPResponse			"Failed"
//	@Failure		500		{object}	helpers.BaseHTTPResponse			"Failed"
//	@Router			/v1/users/register [post]
func (uh UsersHandler) RegisterByUsername(c *gin.Context) {
	req := new(dto.RegisterUserByUsernameRequest)
	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.
			AbortWithStatusJSON(
				helpers.TranslateErrorToStatusCode(err),
				helpers.GenerateBaseResponseWithValidationErrors(
					nil,
					false,
					-1,
					err,
				),
			)
		return
	}

	err = uh.userService.RegisterByUsername(req)
	if err != nil {
		c.
			AbortWithStatusJSON(
				helpers.TranslateErrorToStatusCode(err),
				helpers.GenerateBaseResponseWithError(nil,
					false,
					-1,
					err,
				),
			)
		return
	}

	loginReq := new(dto.LoginByUsernameRequest)
	loginReq.Username = req.Username
	loginReq.Password = req.Password
	var token *dto.TokenDetail
	token, err = uh.userService.LoginByUsername(loginReq)
	if err != nil {
		c.
			AbortWithStatusJSON(
				helpers.TranslateErrorToStatusCode(err),
				helpers.GenerateBaseResponseWithError(nil,
					false,
					-1,
					err,
				),
			)
		return
	}

	c.JSON(
		http.StatusCreated,
		helpers.GenerateBaseResponse(
			token,
			true,
			0,
		),
	)
}

// LoginByUsername godoc
//
//	@Summary		Login By Username
//	@Description	Login By Username
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			Request	body		dto.LoginByUsernameRequest	true	"Login by username"
//	@Success		201		{object}	helpers.BaseHTTPResponse	"Success"
//	@Failure		400		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		404		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		409		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		500		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Router			/v1/users/login [post]
func (uh UsersHandler) LoginByUsername(c *gin.Context) {
	req := new(dto.LoginByUsernameRequest)
	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.
			AbortWithStatusJSON(
				http.StatusBadRequest,
				helpers.GenerateBaseResponseWithValidationErrors(
					nil,
					false,
					-1,
					err,
				),
			)
		return
	}

	var token *dto.TokenDetail
	token, err = uh.userService.LoginByUsername(req)
	if err != nil {
		c.
			AbortWithStatusJSON(
				helpers.TranslateErrorToStatusCode(err),
				helpers.GenerateBaseResponseWithError(nil,
					false,
					-1,
					err,
				),
			)
		return
	}

	c.JSON(
		http.StatusOK,
		helpers.GenerateBaseResponse(
			token,
			true,
			0,
		),
	)
}

// RefreshToken godoc
//
//	@Summary		Refresh access token
//	@Description	Generates a new access token and refresh token using a valid refresh token.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		dto.RefreshTokenRequest		true	"Refresh token"
//
//	@Success		201		{object}	helpers.BaseHTTPResponse	"Success"
//	@Failure		400		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		404		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		409		{object}	helpers.BaseHTTPResponse	"Failed"
//	@Failure		500		{object}	helpers.BaseHTTPResponse	"Failed"
//
//	@Router			/v1/users/refresh-token [post]
func (uh UsersHandler) RefreshToken(c *gin.Context) {
	req := new(dto.RefreshTokenRequest)
	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.
			AbortWithStatusJSON(
				helpers.TranslateErrorToStatusCode(err),
				helpers.GenerateBaseResponseWithValidationErrors(
					nil,
					false,
					-1,
					err,
				),
			)
		return
	}

	var token *dto.TokenDetail
	token, err = uh.userService.RefreshUserToken(req)
	if err != nil {
		c.
			AbortWithStatusJSON(
				helpers.TranslateErrorToStatusCode(err),
				helpers.GenerateBaseResponseWithError(nil,
					false,
					-1,
					err,
				),
			)
		return
	}

	c.JSON(
		http.StatusOK,
		helpers.GenerateBaseResponse(
			token,
			true,
			0,
		),
	)
}
