package myredis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const redisPingTimeout = 5 * time.Second

var ErrNoSuchKey = errors.New("no such key")

type RedisCacher struct {
	rdb *redis.Client
}

func NewCacher(rdb *redis.Client) *RedisCacher {
	return &RedisCacher{
		rdb: rdb,
	}
}

func InitRedis(redisAddr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return rdb, nil
}

func CloseRedis(rdb *redis.Client) {
	if rdb != nil {
		rdb.Close()
	}
}

func serialize(v any) (string, error) {
	data, err := json.Marshal(v)
	return string(data), err
}

func deserialize[T any](data string) (T, error) {
	var result T
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}
