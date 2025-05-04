package config

import (
	"fmt"
	"github.com/alecthomas/kong"
	"os"
	"sherry.archive.com/shared/configs"
	"sherry.archive.com/shared/logger"
)

type Application struct {
	Migrate struct {
		Command string `kong:"arg,name:'command',enum:'up,create'"`
		Option  string `kong:"arg,optional,name:'option'"`
	} `kong:"cmd"`

	AppMode      string               `name:"app-mode" help:"App mode" env:"APP_MODE" default:"production"`
	LoggerConfig logger.Configuration `kong:"help:'Logger config',embed"`

	HTTPPort int `env:"HTTP_PORT" default:"8081"`
	GRPCPort int `env:"GRPC_PORT" default:"8091"`

	MysqlConfig     configs.MysqlConfig `kong:"embed"`
	MigrationFolder string              `env:"MIGRATION_FOLDER" default:"/applications/roulette/indexer/migrations"`

	KafkaConfig  configs.KafkaConfig `kong:"embed"`
	RedisAddress string              `env:"REDIS_ADDRESS" default:"localhost:6379"`
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

func (c *Application) GetPosixByAbsoluteSlot(absSlot uint64) int64 {
	return c.Network.SlotZeroTime + (int64(absSlot)-c.Network.SlotZero)*c.Network.SlotLength
}

func (c *Application) GetPosixTimeBySlot(slot uint64) int64 {
	return c.GetPosixByAbsoluteSlot(slot) / c.Network.SlotLength
}
