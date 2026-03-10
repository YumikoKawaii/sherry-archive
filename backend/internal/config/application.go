package config

import (
	"fmt"
	"os"
)

// Application holds all runtime configuration for the server.
type Application struct {
	Server *ServerConfig `json:"server" mapstructure:"server" yaml:"server"`
	DB     *DBConfig     `json:"db"     mapstructure:"db"     yaml:"db"`
	JWT    *JWTConfig    `json:"jwt"    mapstructure:"jwt"    yaml:"jwt"`
	S3     *S3Config     `json:"s3"     mapstructure:"s3"     yaml:"s3"`
	Redis  *RedisConfig  `json:"redis"  mapstructure:"redis"  yaml:"redis"`
}

type ServerConfig struct {
	Port string `json:"port" mapstructure:"port" yaml:"port"`
}

type DBConfig struct {
	Host     string `json:"host"              mapstructure:"host"              yaml:"host"`
	Port     string `json:"port"              mapstructure:"port"              yaml:"port"`
	User     string `json:"user"              mapstructure:"user"              yaml:"user"`
	Password string `json:"password"          mapstructure:"password"          yaml:"password"`
	Name     string `json:"name"              mapstructure:"name"              yaml:"name"`
	SSLMode  string `json:"ssl_mode"          mapstructure:"ssl_mode"          yaml:"ssl_mode"`
	// MigrationsSource is the golang-migrate source URL.
	// Override via DB__MIGRATIONS_SOURCE, e.g. "file:///app/migrations" in a container
	// or "file:///absolute/path/to/backend/migrations" locally.
	MigrationsSource string `json:"migrations_source" mapstructure:"migrations_source" yaml:"migrations_source"`
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// JWTConfig holds JWT signing secrets and expiry durations.
// Expiry values are parsed as time.Duration strings (e.g. "15m", "168h").
// Env vars use __ as separator: JWT__ACCESS_SECRET, JWT__ACCESS_TOKEN_EXPIRY, etc.
type JWTConfig struct {
	AccessSecret       string `json:"access_secret"        mapstructure:"access_secret"        yaml:"access_secret"`
	RefreshSecret      string `json:"refresh_secret"       mapstructure:"refresh_secret"       yaml:"refresh_secret"`
	AccessTokenExpiry  string `json:"access_token_expiry"  mapstructure:"access_token_expiry"  yaml:"access_token_expiry"`
	RefreshTokenExpiry string `json:"refresh_token_expiry" mapstructure:"refresh_token_expiry" yaml:"refresh_token_expiry"`
}

// S3Config holds AWS S3 connection details.
// PresignExpiry is a time.Duration string (e.g. "1h").
// Credentials are resolved automatically via IAM role (EC2) or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY env vars (local dev).
// Env vars: S3__REGION, S3__BUCKET, S3__PRESIGN_EXPIRY, S3__ENDPOINT (optional, for local MinIO)
type S3Config struct {
	Region        string `json:"region"         mapstructure:"region"         yaml:"region"`
	Bucket        string `json:"bucket"         mapstructure:"bucket"         yaml:"bucket"`
	PresignExpiry string `json:"presign_expiry" mapstructure:"presign_expiry" yaml:"presign_expiry"`
	// Endpoint is optional — set to MinIO URL for local dev (e.g. "http://localhost:9000")
	Endpoint string `json:"endpoint" mapstructure:"endpoint" yaml:"endpoint"`
}

// RedisConfig holds Redis connection details.
// Env vars: REDIS__ADDR, REDIS__PASSWORD, REDIS__DB
type RedisConfig struct {
	Addr     string `json:"addr"     mapstructure:"addr"     yaml:"addr"`
	Password string `json:"password" mapstructure:"password" yaml:"password"`
	DB       int    `json:"db"       mapstructure:"db"       yaml:"db"`
}

func loadDefault() *Application {
	return &Application{
		Server: &ServerConfig{
			Port: "8080",
		},
		DB: &DBConfig{
			Host:             "localhost",
			Port:             "5432",
			User:             "postgres",
			Password:         "postgres",
			Name:             "sherry_archive",
			SSLMode:          "disable",
			MigrationsSource: "/backend/migrations",
		},
		JWT: &JWTConfig{
			AccessSecret:       "change-me-access-secret",
			RefreshSecret:      "change-me-refresh-secret",
			AccessTokenExpiry:  "15m",
			RefreshTokenExpiry: "168h",
		},
		S3: &S3Config{
			Region:        "ap-southeast-1",
			Bucket:        "sherry-archive",
			PresignExpiry: "1h",
			Endpoint:      "http://localhost:9000",
		},
		Redis: &RedisConfig{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
	}
}

// GetMigrationFolder returns the full path to the migration folder
func (a *Application) GetMigrationFolder() string {
	cwd, _ := os.Getwd()
	return fmt.Sprintf("file://%s%s", cwd, a.DB.MigrationsSource)
}
