package main

import (
	"sherry.archive.com/applications/iam/config"
	"sherry.archive.com/applications/iam/servers"
	"sherry.archive.com/shared/logger"
	"sherry.archive.com/shared/migrator"
)

func main() {

	cfg, kongCtx := config.Initialize()
	_, err := logger.InitLogger(cfg.LoggerConfig, logger.LoggerBackendZap)
	if err != nil {
		panic(err)
	}

	switch kongCtx.Command() {
	case "serve":
		servers.Serve(cfg)
	case "migrate <command>":
		switch cfg.Migrate.Command {
		case "up":
			migrator.Up(cfg.MysqlConfig.MigrationDSN(), cfg.GetMigrationFolder())
		}
	case "migrate <command> <option>":
		switch cfg.Migrate.Command {
		case "create":
			migrator.New(cfg.GetMigrationFolder(), cfg.Migrate.Option)
		}
	default:
		panic("unexpected command")
	}
}
