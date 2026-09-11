package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Redis    RedisConfig
	Postgres PostgresConfig
	Logger   LoggerConfig
}

type ServerConfig struct {
	InternalPort int
}

type RedisConfig struct {
	Host               string
	Port               string
	Password           string
	Database           int
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	PoolSize           int
	PoolTimeout        time.Duration
	IdleCheckFrequency time.Duration
	IdleTimeout        time.Duration
}

type PostgresConfig struct {
	Host                  string
	Port                  string
	Username              string
	Password              string
	DbName                string
	SSLMode               string
	MaxIdleConnections    int
	MaxOpenConnections    int
	ConnectionMaxLifetime time.Duration
}

type LoggerConfig struct {
	AppName    string
	FilePath   string
	Level      string
	LoggerName string
}

var (
	configInstance *Config
	once           sync.Once
)

// GetConfig implementing singleton for config
func GetConfig() *Config {
	once.Do(func() {
		configInstance = loadConfig()
	})

	return configInstance
}

func loadConfig() *Config {

	log.Println("Loading config")

	configFilePath, err := getConfigFilePath()
	if err != nil {
		log.Fatalln(err.Error())
	}

	viperConfigInstance, err := readConfigFile(configFilePath)
	if err != nil {
		log.Fatalln(err.Error())
	}

	cfg, err := parseConfig(viperConfigInstance)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return cfg
}

func parseConfig(v *viper.Viper) (*Config, error) {
	cfg := &Config{}
	err := v.Unmarshal(cfg)
	if err != nil {
		log.Println("Error parsing config file")
		return nil, err
	}
	return cfg, nil
}

func readConfigFile(filePath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(filePath)
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		log.Println("Error reading config file")
		return nil, err
	}
	return v, nil
}

func getConfigFilePath() (string, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		log.Println("Error getting APP_ENV")
		return "", errors.New("APP_ENV env variable not set")
	}

	fileType := os.Getenv("APP_CONFIG_FILE_TYPE")
	if fileType == "" {
		log.Println("Error getting APP_CONFIG_FILE_TYPE")
		return "", errors.New("APP_CONFIG_FILE_TYPE env variable not set")
	}

	if env == "docker" {
		return fmt.Sprintf("/app/config/config-%s.%s", env, fileType), nil
	}
	return fmt.Sprintf("./config/config-%s.%s", env, fileType), nil
}
