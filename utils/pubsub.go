package utils

import (
	"github.com/fzxiao233/Vtb_Record/config"
	"github.com/go-redis/redis"
)

var RedisClient *redis.Client

func init() {
	RedisClient = initRedis()
}
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
	_ = RedisClient.Publish(channel, data)
}
