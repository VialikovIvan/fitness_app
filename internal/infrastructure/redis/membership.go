package myredis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	checkoutURLTTL = 24 * time.Hour
)

func (r *RedisCacher) GetCheckoutURL(ctx context.Context, idx string) (string, error) {
	key := "checkout:" + idx

	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNoSuchKey
		}
		return "", err
	}

	return deserialize[string](val)
}

func (r *RedisCacher) SetCheckoutURL(ctx context.Context, idx string, url string) error {
	key := "checkout:" + idx

	serialized, err := serialize(url)
	if err != nil {
		return err
	}

	err = r.rdb.Set(ctx, key, serialized, checkoutURLTTL).Err()

	return err
}

func (r *RedisCacher) DeleteCheckoutURL(ctx context.Context, idx string) error {
	key := "checkout:" + idx

	_, err := r.rdb.Del(ctx, key).Result()
	return err
}

func (r *RedisCacher) EventIsProcessed(ctx context.Context, eventID string) (bool, error) {
	key := "dedup:" + eventID

	_, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *RedisCacher) MarkEventProcessed(ctx context.Context, eventID string) (error) {
	key := "dedup:" + eventID
	_, err := r.rdb.Set(ctx, key, 1, 7*24*time.Hour).Result()
	return err
}
