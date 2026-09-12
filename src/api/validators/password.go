package validators

import (
	"github.com/go-playground/validator/v10"
	"github.com/mahdipeydai/taskmanager-go/common"
	"github.com/mahdipeydai/taskmanager-go/config"
)

func UserPasswordValidator(fld validator.FieldLevel) bool {

	value, ok := fld.Field().Interface().(string)
	if !ok {
		return false
	}
	cfg := config.GetConfig()
	return common.CheckPassword(
		value,
		cfg.Password.MinLength,
		cfg.Password.MaxLength,
		cfg.Password.IncludeChars,
		cfg.Password.IncludeDigits,
		cfg.Password.IncludeLowercase,
		cfg.Password.IncludeUppercase,
	)
}
