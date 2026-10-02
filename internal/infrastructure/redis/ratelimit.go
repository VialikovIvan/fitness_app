package myredis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	tokenBucketCapacity            float64 = 100
	tokenBucketRefillRatePerSecond float64 = 10
	tokenBucketTTL                         = time.Second * 20
)

type TokenBucket struct {
	Tokens     float64
	LastRefill time.Time
}

func newTokenBucket() *TokenBucket {
	return &TokenBucket{
		Tokens:     tokenBucketCapacity,
		LastRefill: time.Now(),
	}
}

func (r *RedisCacher) AllowRequestByIP(ctx context.Context, ip string) (bool, error) {
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

func (r *RedisCacher) getTokenBucketByIP(ctx context.Context, ip string) (*TokenBucket, error) {
	key := "ratelimit:" + ip

	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			tb := newTokenBucket()

			serialized, err := serialize(tb)
			if err != nil {
				return nil, err
			}

			err = r.rdb.Set(ctx, key, serialized, tokenBucketTTL).Err()

			return tb, err
		}
		return nil, err
	}

	return deserialize[*TokenBucket](val)
}

func (r *RedisCacher) tryConsume(tb *TokenBucket) bool {
	elapsedSeconds := time.Since(tb.LastRefill).Seconds()
	tb.Tokens = min(tb.Tokens+elapsedSeconds*tokenBucketRefillRatePerSecond, tokenBucketCapacity)
	tb.LastRefill = time.Now()

	if tb.Tokens >= 1 {
		tb.Tokens -= 1
		return true
	}

	return false
}

func (r *RedisCacher) setTokenBucketByIP(ctx context.Context, ip string, tb *TokenBucket) error {
	key := "ratelimit:" + ip

	serialized, err := serialize(tb)
	if err != nil {
		return err
	}

	err = r.rdb.Set(ctx, key, serialized, tokenBucketTTL).Err()

	return err
}
