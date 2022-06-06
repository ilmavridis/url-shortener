package redisStorage

import (
	"context"

	"ilmavridis/url-shortener/config"

	"github.com/go-redis/redis/v8"
)

var Ctx = context.Background()
var redisClient *redis.Client

func CreateClient() error {

	dbConf := config.Get()
	redisConf := dbConf.Redis

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisConf.Address,
		Password: redisConf.Pass,
		DB:       redisConf.Database,
	})

	// Tests connection
	_, err := redisClient.Ping(context.Background()).Result()

	return err
}

// Returns the connection created
func Get() *redis.Client {
	return redisClient
}
