package postgres

import (
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", dsn)
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

// ConnectXRay opens a PostgreSQL connection instrumented with AWS X-Ray.
// Each query becomes an X-Ray subsegment when a segment is active in ctx.
func ConnectXRay(dsn string) (*sqlx.DB, error) {
	rawDB, err := xray.SQLContext("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// sqlx.NewDb tells sqlx to use $N bind vars for the postgres dialect,
	// regardless of the internal driver name xray.SQL registers.
	db := sqlx.NewDb(rawDB, "postgres")
	if err := db.Ping(); err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return db, nil
}
