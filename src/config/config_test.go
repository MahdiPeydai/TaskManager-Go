package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestGetConfigFilePath(t *testing.T) {
	tests := []struct {
		name        string
		appEnv      string
		fileType    string
		expected    string
		expectError bool
	}{
		{
			name:     "local development",
			appEnv:   "dev",
			fileType: "yaml",
			expected: "./config/config-dev.yaml",
		},
		{
			name:     "test environment",
			appEnv:   "test",
			fileType: "yaml",
			expected: "./config/config-test.yaml",
		},
		{
			name:     "docker environment",
			appEnv:   "docker",
			fileType: "yaml",
			expected: "/app/config/config-docker.yaml",
		},
		{
			name:        "missing APP_ENV",
			appEnv:      "",
			fileType:    "yaml",
			expectError: true,
		},
		{
			name:        "missing APP_CONFIG_FILE_TYPE",
			appEnv:      "dev",
			fileType:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.appEnv)
			t.Setenv("APP_CONFIG_FILE_TYPE", tt.fileType)

			got, err := getConfigFilePath()

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	v := viper.New()

	v.Set("redis.host", "localhost")
	v.Set("redis.port", "6379")
	v.Set("redis.password", "secret")
	v.Set("redis.database", 1)
	v.Set("redis.poolSize", 10)

	v.Set("postgres.host", "localhost")
	v.Set("postgres.port", "5432")
	v.Set("postgres.username", "postgres")
	v.Set("postgres.password", "secret")
	v.Set("postgres.dbName", "taskmanager")

	cfg, err := parseConfig(v)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Redis.Host != "localhost" {
		t.Errorf("expected Redis host localhost, got %q", cfg.Redis.Host)
	}

	if cfg.Redis.Port != "6379" {
		t.Errorf("expected Redis port 6379, got %q", cfg.Redis.Port)
	}

	if cfg.Redis.Database != 1 {
		t.Errorf("expected Redis database 1, got %d", cfg.Redis.Database)
	}

	if cfg.Redis.PoolSize != 10 {
		t.Errorf("expected Redis pool size 10, got %d", cfg.Redis.PoolSize)
	}

	if cfg.Postgres.Host != "localhost" {
		t.Errorf("expected Postgres host localhost, got %q", cfg.Postgres.Host)
	}

	if cfg.Postgres.DbName != "taskmanager" {
		t.Errorf("expected database name taskmanager, got %q", cfg.Postgres.DbName)
	}
}

func TestReadConfigFile(t *testing.T) {
	configContent := `
redis:
  host: localhost
  port: "6379"
  database: 1

postgres:
  host: localhost
  port: "5432"
  username: postgres
  dbName: taskmanager
`

	file := filepath.Join(t.TempDir(), "config.yaml")

	err := os.WriteFile(
		file,
		[]byte(configContent),
		0644,
	)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	v, err := readConfigFile(file)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if v.GetString("redis.host") != "localhost" {
		t.Errorf(
			"expected localhost, got %q",
			v.GetString("redis.host"),
		)
	}

	if v.GetString("postgres.username") != "postgres" {
		t.Errorf(
			"expected postgres, got %q",
			v.GetString("postgres.username"),
		)
	}
}

func TestReadConfigFile_FileNotFound(t *testing.T) {
	_, err := readConfigFile("/does/not/exist/config.yaml")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadConfigFile_InvalidYAML(t *testing.T) {
	configContent := `
redis:
  host: localhost
  port: [invalid
`

	file := filepath.Join(t.TempDir(), "config.yaml")

	err := os.WriteFile(file, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	_, err = readConfigFile(file)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
