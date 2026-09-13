package mocks

import (
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

var _ logging.LoggerInterface = (*MockLogger)(nil)

func (m *MockLogger) Init() {}

func (m *MockLogger) Debug(
	cat logging.LogCategory,
	sub logging.LogSubCategory,
	msg string,
	extras map[logging.ExtraKey]interface{},
) {
}

func (m *MockLogger) Debugf(template string, args ...interface{}) {}

func (m *MockLogger) Info(
	cat logging.LogCategory,
	sub logging.LogSubCategory,
	msg string,
	extras map[logging.ExtraKey]interface{},
) {
}

func (m *MockLogger) Infof(template string, args ...interface{}) {}

func (m *MockLogger) Warn(
	cat logging.LogCategory,
	sub logging.LogSubCategory,
	msg string,
	extras map[logging.ExtraKey]interface{},
) {
}

func (m *MockLogger) Warnf(template string, args ...interface{}) {}

func (m *MockLogger) Error(
	cat logging.LogCategory,
	sub logging.LogSubCategory,
	msg string,
	extras map[logging.ExtraKey]interface{},
) {
}

func (m *MockLogger) Errorf(template string, args ...interface{}) {}

func (m *MockLogger) Fatal(
	cat logging.LogCategory,
	sub logging.LogSubCategory,
	msg string,
	extras map[logging.ExtraKey]interface{},
) {
}

func (m *MockLogger) Fatalf(template string, args ...interface{}) {}
