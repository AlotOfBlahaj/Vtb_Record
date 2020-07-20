package utils

import (
	"context"
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/go-redis/redis"
)

var RedisClient *redis.Client

func initRedis() *redis.Client {
	RedisClient := redis.NewClient(
		&redis.Options{
			Addr:     config.Config.RedisHost,
			Password: "",
			DB:       0,
		})
	return RedisClient
}
func Publish(data []byte, channel string) {
	if RedisClient == nil {
		RedisClient = initRedis()
	}
	_ = RedisClient.Publish(context.Background(), channel, data)
}
