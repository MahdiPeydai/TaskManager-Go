package mocks

import (
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type MockDatabase struct {
	DB   *gorm.DB
	Mock sqlmock.Sqlmock
}

func NewMockDatabase() (*MockDatabase, error) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, err
	}

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		sqlDB.Close()
		return nil, err
	}

	return &MockDatabase{
		DB:   db,
		Mock: mock,
	}, nil
}

func ExpectQuery(query string) string {
	return regexp.QuoteMeta(query)
}
