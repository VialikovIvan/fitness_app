package myredis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	accessTTL = 5 * time.Minute
)

func (r *RedisCacher) GetAccessURL(ctx context.Context, userID string) (string, error) {
	key := "access:" + userID

	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNoSuchKey
		}
		return "", err
	}

	return deserialize[string](val)
}

func (r *RedisCacher) SetAccessURL(ctx context.Context, userID string, url string) error {
	key := "access:" + userID

	serialized, err := serialize(url)
	if err != nil {
		return err
	}

	return r.rdb.Set(ctx, key, serialized, accessTTL).Err()
}

func (r *RedisCacher) DeleteAccessURL(ctx context.Context, userID string) error {
	key := "access:" + userID
	return r.rdb.Del(ctx, key).Err()
}

func (r *RedisCacher) MarkJTIUsed(ctx context.Context, jti string) (bool, error) {
    key := "jti:" + jti
    return r.rdb.SetNX(ctx, key, "1", accessTTL).Result()
}

func (r *RedisCacher) SetAccessCooldown(ctx context.Context, userID string) error {
	key := "access-cooldown:" + userID
	return r.rdb.Set(ctx, key, "1", accessTTL).Err()
}

func (r *RedisCacher) IsInCooldown(ctx context.Context, userID string) (bool, error) {
	key := "access-cooldown:" + userID

	exists, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return true, err
	}

	return exists >= 1, nil
}
