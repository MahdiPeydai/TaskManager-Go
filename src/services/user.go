package services

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/mahdipeydai/taskmanager-go/api/dto"
	"github.com/mahdipeydai/taskmanager-go/common"
	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/constants"
	"github.com/mahdipeydai/taskmanager-go/data/models"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"github.com/mahdipeydai/taskmanager-go/pkg/service_errors"
	"gorm.io/gorm"
)

type UsersService struct {
	logger       logging.LoggerInterface
	cfg          *config.Config
	tokenService *TokenService
	database     *gorm.DB
}

func GetUsersService(database *gorm.DB, cfg *config.Config, logger logging.LoggerInterface) *UsersService {
	return &UsersService{
		logger:       logger,
		cfg:          cfg,
		tokenService: GetTokenService(cfg, logger),
		database:     database,
	}
}

func (s *UsersService) CreateUserToken(user *models.User) (*dto.TokenDetail, error) {
	td := Token{
		UserId:       user.Id,
		Username:     user.Username,
		FirstName:    user.Firstname,
		LastName:     user.Lastname,
		Email:        user.Email,
		MobileNumber: user.MobileNumber,
	}

	for _, ur := range user.Roles {
		td.Roles = append(td.Roles, ur.Role.Name)
	}
	token, err := s.tokenService.GenerateToken(td)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *UsersService) RefreshUserToken(req *dto.RefreshTokenRequest) (*dto.TokenDetail, error) {
	verifiedToken, err := s.tokenService.VerifyToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok || !verifiedToken.Valid {
		return nil, service_errors.ServiceError{
			EndUserMessage: service_errors.InvalidToken,
		}
	}

	tokenType, ok := claims[constants.TokenTypeKey].(string)
	if !ok || tokenType != constants.RefreshToken {
		return nil, service_errors.ServiceError{
			EndUserMessage: service_errors.InvalidToken,
		}
	}

	userID, ok := claims[constants.UserIdKey].(float64)
	if !ok {
		return nil, service_errors.ServiceError{
			EndUserMessage: service_errors.InvalidToken,
		}
	}

	var u models.User
	err = s.database.
		Model(&models.User{}).
		Where("id = ?", userID).
		Preload("Roles", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("Role")
		}).
		First(&u).Error

	if err != nil {
		return nil, service_errors.ServiceError{EndUserMessage: service_errors.RecordNotFound}
	}

	return s.CreateUserToken(&u)
}

func (s *UsersService) RegisterByUsername(req *dto.RegisterUserByUsernameRequest) error {
	exists, err := s.existsByUsername(req.Username)
	if err != nil {
		return err
	}
	if exists {
		return service_errors.ServiceError{EndUserMessage: service_errors.UsernameExists}
	}

	exists, err = s.existsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return service_errors.ServiceError{EndUserMessage: service_errors.EmailExists}
	}

	u := models.User{
		Username:  req.Username,
		Firstname: &req.FirstName,
		Lastname:  &req.LastName,
		Email:     &req.Email,
	}
	var password string
	password, err = common.HashPassword(req.Password)
	if err != nil {
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}
		s.logger.Error(logging.General, logging.HashPassword, "Failed to hash password", extras)
		return err
	}

	u.Password = &password

	roleId, err := s.getDefaultRole()
	if err != nil {
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}
		s.logger.Error(logging.General, logging.HashPassword, "Failed to get default role", extras)
		return err
	}

	tx := s.database.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	err = tx.Create(&u).Error
	if err != nil {
		tx.Rollback()
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}
		s.logger.Error(logging.General, logging.HashPassword, "Failed to create user", extras)
		return err
	}

	err = tx.Create(&models.UserRole{RoleId: roleId, UserId: u.Id}).Error
	if err != nil {
		tx.Rollback()
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}
		s.logger.Error(logging.General, logging.HashPassword, "Failed to create user role", extras)
		return err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		extras := map[logging.ExtraKey]interface{}{
			logging.ErrorMessage: err.Error(),
		}
		s.logger.Error(logging.General, logging.HashPassword, "Failed to create user", extras)
		return err
	}
	return nil
}

func (s *UsersService) LoginByUsername(req *dto.LoginByUsernameRequest) (*dto.TokenDetail, error) {
	var u models.User
	err := s.database.
		Model(&models.User{}).
		Where("username = ?", req.Username).
		Preload("Roles", func(tx *gorm.DB) *gorm.DB {
			return tx.Preload("Role")
		}).
		First(&u).Error

	if err != nil {
		return nil, service_errors.ServiceError{
			EndUserMessage: service_errors.RecordNotFound,
		}
	}

	if u.Password == nil || !common.ComparePasswords(*u.Password, req.Password) {
		return nil, service_errors.ServiceError{
			EndUserMessage: service_errors.RecordNotFound,
		}
	}

	return s.CreateUserToken(&u)
}

func (s *UsersService) existsByEmail(email string) (bool, error) {
	var exists bool
	if err := s.database.Model(&models.User{}).
		Select("count(*) > 0").
		Where("email = ?", email).
		Find(&exists).
		Error; err != nil {
		s.logger.Error(logging.Postgres, logging.Select, err.Error(), nil)
		return false, err
	}
	return exists, nil
}

func (s *UsersService) existsByUsername(username string) (bool, error) {
	var exists bool
	if err := s.database.Model(&models.User{}).
		Select("count(*) > 0").
		Where("username = ?", username).
		Find(&exists).
		Error; err != nil {
		s.logger.Error(logging.Postgres, logging.Select, err.Error(), nil)
		return false, err
	}
	return exists, nil
}

func (s *UsersService) getDefaultRole() (roleId int, err error) {
	if err = s.database.Model(&models.Role{}).
		Select("id").
		Where("name = ?", constants.DefaultRoleName).
		First(&roleId).Error; err != nil {
		return 0, err
	}
	return roleId, nil
}
