package logging

type LogCategory string
type LogSubCategory string
type ExtraKey string

const (
	General         LogCategory = "General"
	Internal        LogCategory = "Internal"
	Postgres        LogCategory = "Postgres"
	Redis           LogCategory = "Redis"
	Prometheus      LogCategory = "Prometheus"
	Opentelemetry   LogCategory = "Opentelemetry"
	RequestResponse LogCategory = "RequestResponse"
)

const (
	// general
	StartUp LogSubCategory = "StartUp"
	Close   LogSubCategory = "Close"
	//ExternalService LogSubCategory = "ExternalService"

	// internal
	API          LogSubCategory = "API"
	HashPassword LogSubCategory = "HashPassword"
	//DefaultRoleNotFound LogSubCategory = "DefaultRoleNotFound"

	// postgres
	Migration LogSubCategory = "Migration"
	Select    LogSubCategory = "Select"
	Rollback  LogSubCategory = "Rollback"
	Update    LogSubCategory = "Update"
	Delete    LogSubCategory = "Delete"
	Insert    LogSubCategory = "Insert"

	// Internal
	CloseFile LogSubCategory = "Close File"
)

const (
	ClientIp     ExtraKey = "ClientIp"
	Method       ExtraKey = "Method"
	StatusCode   ExtraKey = "StatusCode"
	BodySize     ExtraKey = "BodySize"
	Path         ExtraKey = "Path"
	Latency      ExtraKey = "Latency"
	ResponseBody ExtraKey = "ResponseBody"
	ErrorMessage ExtraKey = "ErrorMessage"
)
