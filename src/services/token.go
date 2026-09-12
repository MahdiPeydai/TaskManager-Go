package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"github.com/mahdipeydai/taskmanager-go/pkg/service_errors"
)

type TokenService struct {
	logger logging.LoggerInterface
	cfg    *config.Config
}

type Token struct {
	UserId       int
	FirstName    *string
	LastName     *string
	Username     string
	MobileNumber *string
	Email        *string
	Roles        []string
}

func GetTokenService(cfg *config.Config, logger logging.LoggerInterface) *TokenService {
	return &TokenService{
		logger: logger,
		cfg:    cfg,
	}
}

func (s *TokenService) GenerateToken(token Token) (*dto.TokenDetail, error) {
	td := &dto.TokenDetail{}

	td.AccessTokenExpireTime = time.Now().Add(s.cfg.Jwt.AccessTokenExpireTime * time.Second).Unix()
	td.RefreshTokenExpireTime = time.Now().Add(s.cfg.Jwt.RefreshTokenExpireTime * time.Second).Unix()

	atc := jwt.MapClaims{}

	atc[constants.UserIdKey] = token.UserId
	atc[constants.FirstNameKey] = token.FirstName
	atc[constants.LastNameKey] = token.LastName
	atc[constants.UsernameKey] = token.Username
	atc[constants.MobileNumberKey] = token.MobileNumber
	atc[constants.EmailKey] = token.Email
	atc[constants.RolesKey] = token.Roles
	atc[constants.ExpireTimeKey] = td.AccessTokenExpireTime
	atc[constants.TokenTypeKey] = constants.AccessToken

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atc)

	signedAccessToken, err := at.SignedString([]byte(s.cfg.Jwt.Secret))
	if err != nil {
		return nil, err
	}
	td.AccessToken = fmt.Sprintf("%s %s", constants.AuthorizationHeaderPrefix, signedAccessToken)

	rtc := jwt.MapClaims{}

	rtc[constants.UserIdKey] = token.UserId
	rtc[constants.ExpireTimeKey] = td.RefreshTokenExpireTime
	rtc[constants.TokenTypeKey] = constants.RefreshToken

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtc)
	td.RefreshToken, err = rt.SignedString([]byte(s.cfg.Jwt.Secret))

	if err != nil {
		return nil, err
	}

	return td, nil
}

func (s *TokenService) VerifyToken(token string) (*jwt.Token, error) {
	at, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, service_errors.ServiceError{EndUserMessage: service_errors.UnexpectedError}
		}
		return []byte(s.cfg.Jwt.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	return at, nil
}

func (s *TokenService) GetClaims(token string) (claimMap map[string]interface{}, err error) {
	claimMap = make(map[string]interface{})

	verifiedToken, err := s.VerifyToken(token)
	if err != nil {
		return nil, err
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if ok && verifiedToken.Valid {
		for key, value := range claims {
			claimMap[key] = value
		}
		return claimMap, nil
	}
	return nil, service_errors.ServiceError{EndUserMessage: service_errors.ClaimsNotFound}
}
