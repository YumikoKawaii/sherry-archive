package migrator

import (
	"errors"
	"fmt"
	"os"
	"sherry.archive.com/shared/logger"
	"strings"
	"time"

	migrateV4 "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const versionTimeFormat = "20060102150405"

// New with up and down options.
func New(migrationFolder string, name string) {
	folder := strings.ReplaceAll(migrationFolder, "file://", "")
	now := time.Now()
	ver := now.Format(versionTimeFormat)

	up := fmt.Sprintf("%s/%s_%s.up.sql", folder, ver, name)

	logger.Infof("create migrate: %s", name)

	if err := os.WriteFile(up, []byte{}, 0644); err != nil {
		logger.Fatalf("create migrate up error: %v", err)
	}
}

// Up migrate db to latest version.
func Up(dsn string, migrationFolder string) {
	m, err := migrateV4.New(migrationFolder, dsn)
	if err != nil {
		logger.Fatalf("error create migrate: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrateV4.ErrNoChange) {
		logger.Fatalf("error when migrate up: %v", err)
	}

	logger.Info("migrate up completed")
}
