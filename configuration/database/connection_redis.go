package database

import (
	"context"
	"fmt"
	"os"

	"github.com/Higor-ViniciusDev/posgo_raterlimite/configuration/logger"
	"github.com/redis/go-redis/v9"
)

func NewConnectionRedis() *redis.Client {
	url := os.Getenv("REDIS_URL")
	port := os.Getenv("REDIS_PORT")

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", url, port),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Error("Falha ao conectar ao Redis", err)
		panic(fmt.Sprintf("Falha ao conectar ao Redis: %v", err))
	}

	return rdb
}
