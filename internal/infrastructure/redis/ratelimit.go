package myredis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	TokenBucketCapacity            float64 = 100
	TokenBucketRefillRatePerSecond float64 = 10
	TokenBucketTTL                         = time.Second * 20
)

type RedisLimiter struct {
	rdb *redis.Client
}

func NewLimiter(rdb *redis.Client) *RedisLimiter {
	return &RedisLimiter{
		rdb: rdb,
	}
}

type TokenBucket struct {
	Tokens     float64
	LastRefill time.Time
}

func newTokenBucket() *TokenBucket {
	return &TokenBucket{
		Tokens:     TokenBucketCapacity,
		LastRefill: time.Now(),
	}
}

func (r *RedisLimiter) AllowRequestByIP(ctx context.Context, ip string) (bool, error) {
	tb, err := r.getTokenBucketByIP(ctx, ip)
	if err != nil {
		return true, err
	}

	// updates tb struct inside
	consumed := r.tryConsume(tb)

	err = r.setTokenBucketByIP(ctx, ip, tb)
	if err != nil {
		return true, err
	}

	if !consumed {
		return false, nil
	}

	return true, nil
}

func (r *RedisLimiter) getTokenBucketByIP(ctx context.Context, ip string) (*TokenBucket, error) {
	key := "ratelimit:" + ip

	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			tb := newTokenBucket()

			serialized, err := serialize(tb)
			if err != nil {
				return nil, err
			}

			err = r.rdb.Set(ctx, key, serialized, TokenBucketTTL).Err()

			return tb, err
		}
		return nil, err
	}

	return deserialize[*TokenBucket](val)
}

func (r *RedisLimiter) tryConsume(tb *TokenBucket) bool {
	elapsedSeconds := time.Since(tb.LastRefill).Seconds()
	tb.Tokens = min(tb.Tokens+elapsedSeconds*TokenBucketRefillRatePerSecond, TokenBucketCapacity)
	tb.LastRefill = time.Now()

	if tb.Tokens >= 1 {
		tb.Tokens -= 1
		return true
	}

	return false
}

func (r *RedisLimiter) setTokenBucketByIP(ctx context.Context, ip string, tb *TokenBucket) error {
	key := "ratelimit:" + ip

	serialized, err := serialize(tb)
	if err != nil {
		return err
	}

	err = r.rdb.Set(ctx, key, serialized, TokenBucketTTL).Err()

	return err
}
