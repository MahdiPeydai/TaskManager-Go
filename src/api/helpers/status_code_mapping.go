package helpers

import (
	"net/http"

	"github.com/mahdipeydai/taskmanager-go/pkg/service_errors"
)

var StatusCodeMapping = map[string]int{
	service_errors.UnexpectedError: 500,
	service_errors.ClaimsNotFound:  500,

	service_errors.EmailExists:           409,
	service_errors.UsernameExists:        409,
	service_errors.WrongUsernamePassword: 401,

	service_errors.RecordNotFound: 404,
}

func TranslateErrorToStatusCode(err error) int {
	value, ok := StatusCodeMapping[err.Error()]
	if !ok {
		return http.StatusInternalServerError
	}
	return value
}
