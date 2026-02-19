package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"strings"

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

	source, err := resolveSource(cfg.DB.MigrationsSource)
	if err != nil {
		log.Fatalf("migrations source: %v", err)
	}

	if err := up(db.DB, source); err != nil {
		log.Fatalf("migrate up: %v", err)
	}
	log.Println("migrations applied successfully")
}

// resolveSource converts a relative file:// URL to an absolute file:/// URL.
// golang-migrate requires file:///absolute/path — "file://relative" is parsed
// by Go's net/url as host="relative", path="", which causes a "no scheme" error.
func resolveSource(source string) (string, error) {
	const prefix = "file://"
	if !strings.HasPrefix(source, prefix) {
		return source, nil
	}
	rel := strings.TrimPrefix(source, prefix)
	if filepath.IsAbs(rel) {
		// already absolute, just normalise to file:///path
		return prefix + rel, nil
	}
	abs, err := filepath.Abs(rel)
	if err != nil {
		return "", fmt.Errorf("resolve %q: %w", rel, err)
	}
	return prefix + abs, nil
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
