package migrate

import (
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
	"github.com/yumikokawaii/sherry-archive/internal/config"
	"github.com/yumikokawaii/sherry-archive/internal/repository/postgres"
)

func Up(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := postgres.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	if err := up(db.DB); err != nil {
		log.Fatalf("migrate up: %v", err)
	}
	log.Println("migrations applied successfully")
}

func up(db *sql.DB) error {
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
