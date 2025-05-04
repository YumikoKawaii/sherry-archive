package database

import (
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sherry.archive.com/shared/logger"
	"sync"
)

var (
	mysqlOnce   sync.Once
	mysqlGormDB *gorm.DB
	mysqlDB     *sql.DB
)

func NewMysqlGormDatabase(dsn string) *gorm.DB {
	mysqlOnce.Do(func() {
		_db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		if err != nil {
			panic(err)
		}

		mysqlGormDB = _db
		logger.Info("connect to mysql success")
	})

	return mysqlGormDB
}

func NewMysqlDB(dsn string) *sql.DB {
	mysqlOnce.Do(func() {
		_db, err := sql.Open("mysql", dsn)
		if err != nil {
			panic(err)
		}

		mysqlDB = _db
		logger.Info("connect to mysql success")
	})

	return mysqlDB
}
