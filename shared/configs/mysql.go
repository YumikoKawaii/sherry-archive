package configs

import "fmt"

type MysqlConfig struct {
	Username string `env:"MYSQL_USERNAME"`
	Password string `env:"MYSQL_PASSWORD"`
	Host     string `env:"MYSQL_HOST"`
	Port     int    `env:"MYSQL_PORT"`
	Database string `env:"MYSQL_DATABASE"`
}

func (c *MysqlConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", c.Username, c.Password, c.Host, c.Port, c.Database)
}

func (c *MysqlConfig) MigrationDSN() string {
	return fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?parseTime=true", c.Username, c.Password, c.Host, c.Port, c.Database)
}
