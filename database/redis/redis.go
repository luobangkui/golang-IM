package redis

import (
	"github.com/luobangkui/im-learn-about/config"
	"gopkg.in/redis.v5"
)

func NewRedisClient(config config.Config) *redis.Client {
	db := config.Redis
	return redis.NewClient(&redis.Options{
		Addr:db.Addr,
		DB:db.Db,
		Password:db.Pass,
	})
}