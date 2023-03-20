package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
)

func NewRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return rdb
}

func Publish(data, chanel string) error {
	rd := NewRedis()
	err := rd.Publish(context.Background(), chanel, data).Err()
	return err
}
