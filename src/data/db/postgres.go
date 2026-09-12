package db

import (
	"fmt"
	"time"

	"github.com/mahdipeydai/taskmanager-go/config"
	"github.com/mahdipeydai/taskmanager-go/pkg/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var logger = logging.GetLogger(config.GetConfig())

var dbClient *gorm.DB

func InitDb(cfg *config.Config) error {
	cnn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Username,
		cfg.Postgres.Password,
		cfg.Postgres.DbName,
		cfg.Postgres.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(cnn), &gorm.Config{})

	if err != nil {
		return err
	}

	sqlDb, err := db.DB()
	if err != nil {
		return err
	}

	sqlDb.SetMaxIdleConns(cfg.Postgres.MaxIdleConnections)
	sqlDb.SetMaxOpenConns(cfg.Postgres.MaxOpenConnections)
	sqlDb.SetConnMaxLifetime(cfg.Postgres.ConnectionMaxLifetime * time.Minute)

	logger.Info(logging.Postgres, logging.StartUp, "Database connection established", nil)
	dbClient = db
	return nil
}

func GetDB() *gorm.DB {
	return dbClient
}

func CloseDb() {
	conn, _ := dbClient.DB()
	err := conn.Close()
	if err != nil {
		logger.Fatal(logging.Postgres, logging.Close, err.Error(), nil)
	}
}
