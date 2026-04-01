package migrate

import (
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"github.com/yumikokawaii/sherry-archive/internal/config"
	"github.com/yumikokawaii/sherry-archive/internal/repository/postgres"
)

func Up(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("config", zap.Error(err))
	}

	db, err := postgres.Connect(cfg.DB.DSN())
	if err != nil {
		zap.L().Fatal("db connect", zap.Error(err))
	}
	defer db.Close()

	if err := up(db.DB, cfg.GetMigrationFolder()); err != nil {
		zap.L().Fatal("migrate up", zap.Error(err))
	}
	zap.L().Info("migrations applied successfully")
}

func up(db *sql.DB, source string) error {
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(source, "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
