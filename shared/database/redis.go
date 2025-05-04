package database

import (
	"github.com/go-redis/redis/v8"
	"golang.org/x/xerrors"
	"sync"
)

var (
	redisOnce sync.Once
	redisCli  *redis.Client
)

func NewRedisClient(dsn string) *redis.Client {
	redisOnce.Do(func() {
		_cli := redis.NewClient(&redis.Options{
			Addr: dsn,
		})
		if err := _cli.Ping(_cli.Context()).Err(); err != nil {
			panic(xerrors.Errorf("cannot connect to redis: %s", err.Error()))
		}

		redisCli = _cli
	})

	return redisCli
}
