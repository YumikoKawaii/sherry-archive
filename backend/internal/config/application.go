package config

import "fmt"

// Application holds all runtime configuration for the server.
type Application struct {
	Server *ServerConfig `json:"server" mapstructure:"server" yaml:"server"`
	DB     *DBConfig     `json:"db"     mapstructure:"db"     yaml:"db"`
	JWT    *JWTConfig    `json:"jwt"    mapstructure:"jwt"    yaml:"jwt"`
	MinIO  *MinIOConfig  `json:"minio"  mapstructure:"minio"  yaml:"minio"`
}

type ServerConfig struct {
	Port string `json:"port" mapstructure:"port" yaml:"port"`
}

type DBConfig struct {
	Host             string `json:"host"              mapstructure:"host"              yaml:"host"`
	Port             string `json:"port"              mapstructure:"port"              yaml:"port"`
	User             string `json:"user"              mapstructure:"user"              yaml:"user"`
	Password         string `json:"password"          mapstructure:"password"          yaml:"password"`
	DBName           string `json:"db_name"           mapstructure:"db_name"           yaml:"db_name"`
	SSLMode          string `json:"ssl_mode"          mapstructure:"ssl_mode"          yaml:"ssl_mode"`
	// MigrationsSource is the golang-migrate source URL.
	// Override via DB__MIGRATIONS_SOURCE, e.g. "file:///app/migrations" in a container
	// or "file:///absolute/path/to/backend/migrations" locally.
	MigrationsSource string `json:"migrations_source" mapstructure:"migrations_source" yaml:"migrations_source"`
}

func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
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

// MinIOConfig holds MinIO connection details.
// PresignExpiry is a time.Duration string (e.g. "1h").
// Env vars: MINIO__ENDPOINT, MINIO__ACCESS_KEY_ID, MINIO__SECRET_ACCESS_KEY, etc.
type MinIOConfig struct {
	Endpoint        string `json:"endpoint"          mapstructure:"endpoint"          yaml:"endpoint"`
	AccessKeyID     string `json:"access_key_id"     mapstructure:"access_key_id"     yaml:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key" mapstructure:"secret_access_key" yaml:"secret_access_key"`
	Bucket          string `json:"bucket"            mapstructure:"bucket"            yaml:"bucket"`
	UseSSL          bool   `json:"use_ssl"           mapstructure:"use_ssl"           yaml:"use_ssl"`
	PresignExpiry   string `json:"presign_expiry"    mapstructure:"presign_expiry"    yaml:"presign_expiry"`
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
			DBName:           "sherry_archive",
			SSLMode:          "disable",
			MigrationsSource: "file://migrations",
		},
		JWT: &JWTConfig{
			AccessSecret:       "change-me-access-secret",
			RefreshSecret:      "change-me-refresh-secret",
			AccessTokenExpiry:  "15m",
			RefreshTokenExpiry: "168h",
		},
		MinIO: &MinIOConfig{
			Endpoint:        "localhost:9000",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			Bucket:          "sherry-archive",
			UseSSL:          false,
			PresignExpiry:   "1h",
		},
	}
}
