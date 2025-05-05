package config

import (
	"fmt"
	"github.com/alecthomas/kong"
	"os"
	"sherry.archive.com/applications/archive/adapters/multimedia"
	"sherry.archive.com/shared/configs"
	"sherry.archive.com/shared/logger"
)

type Application struct {
	Serve   struct{} `kong:"cmd"`
	Extract struct{} `kong:"cmd"`
	Migrate struct {
		Command string `kong:"arg,name:'command',enum:'up,create'"`
		Option  string `kong:"arg,optional,name:'option'"`
	} `kong:"cmd"`

	AppMode      string               `name:"app-mode" help:"App mode" env:"APP_MODE" default:"development"`
	LoggerConfig logger.Configuration `kong:"help:'Logger config',embed"`

	HTTPPort int `env:"HTTP_PORT" default:"8081"`
	GRPCPort int `env:"GRPC_PORT" default:"8091"`

	MysqlConfig     configs.MysqlConfig `kong:"embed"`
	MigrationFolder string              `env:"MIGRATION_FOLDER" default:"/applications/archive/migrations"`

	KafkaConfig  configs.KafkaConfig `kong:"embed"`
	RedisAddress string              `env:"REDIS_ADDRESS" default:"localhost:6379"`

	CloudinaryConfig multimedia.CloudinaryConfig `kong:"embed"`
}

func Initialize() (*Application, *kong.Context) {
	cfg := Application{}
	kongCtx := kong.Parse(&cfg)
	return &cfg, kongCtx
}

func (c *Application) GetMigrationFolder() string {
	cwd, _ := os.Getwd()
	return fmt.Sprintf("file://%s%s", cwd, c.MigrationFolder)
}
