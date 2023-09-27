package db

import (
	"context"
	"fmt"
	"log"
	"x-tentioncrew/user-service/pkg/config"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(cfg config.Config) *redis.Client {
	ctx := context.Background()
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0, // use default DB
	})

	responce, redisErr := rdb.Ping(ctx).Result()
	if redisErr != nil {
		log.Fatalln("redis connection failed ", redisErr.Error())
	}

	fmt.Println(responce)
	return rdb
}
