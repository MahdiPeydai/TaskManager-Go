package service_errors

const (
	// Token
	UnexpectedError    = "Unexpected error"
	ClaimsNotFound     = "Claims not found"
	TokenRequired      = "Token is required"
	TokenExpired       = "Token is expired"
	InvalidToken       = "Token is invalid"
	InvalidTokenSchema = "Token is invalid, invalid schema"
	PermissionDenied   = "PermissionDenied"

	// User
	EmailExists           = "Email already exists"
	UsernameExists        = "Username already exists"
	WrongUsernamePassword = "Wrong username or password"
	InvalidUserId         = "Invalid user id"

	// DB
	RecordNotFound = "Record not found"

	// Validation
	InvalidId = "Invalid id"

	// Task
	AssigneePermissionDenied = "Assignee permission denied"
)
