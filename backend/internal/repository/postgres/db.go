package postgres

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(dsn string) (*sqlx.DB, error) {
	return ConnectWithDriver("postgres", dsn)
}

// ConnectWithDriver opens a connection using the given driver name.
// Pass the instrumented driver name from tracing.Init to enable DB span tracing;
// pass "postgres" for the plain lib/pq driver.
func ConnectWithDriver(driverName, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return db, nil
}
