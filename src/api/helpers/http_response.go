package helpers

import "github.com/mahdipeydai/taskmanager-go/api/validators"

type BaseHTTPResponse struct {
	Result           any                          `json:"result"`
	Success          bool                         `json:"success"`
	ResultCode       ResultCode                   `json:"result_code"`
	Error            any                          `json:"error"`
	ValidationErrors []validators.ValidationError `json:"validation_errors"`
}

func GenerateBaseResponse(result any, success bool, resultCode ResultCode) BaseHTTPResponse {
	return BaseHTTPResponse{
		Result:     result,
		Success:    success,
		ResultCode: resultCode,
	}
}

func GenerateBaseResponseWithError(result any, success bool, resultCode ResultCode, err error) BaseHTTPResponse {
	return BaseHTTPResponse{
		Result:     result,
		Success:    success,
		ResultCode: resultCode,
		Error:      err.Error(),
	}
}

func GenerateBaseResponseWithAnyError(result any, success bool, resultCode ResultCode, err any) *BaseHTTPResponse {
	return &BaseHTTPResponse{
		Result:     result,
		Success:    success,
		ResultCode: resultCode,
		Error:      err,
	}
}

func GenerateBaseResponseWithValidationErrors(result any, success bool, resultCode ResultCode, err error) BaseHTTPResponse {
	return BaseHTTPResponse{
		Result:           result,
		Success:          success,
		ResultCode:       resultCode,
		ValidationErrors: validators.GetValidationErrors(err),
	}
}
