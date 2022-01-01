package xredis

import (
	"context"
	"github.com/go-redis/redis/v8"
)

// ScanAll scans all keys from given pattern and scan count, is a wrapper function for SCAN.
func ScanAll(ctx context.Context, client *redis.Client, match string, count int64) (keys []string, err error) {
	cursor := uint64(0)
	tempKeys := make([]string, 0)
	for {
		tempKeys, cursor, err = client.Scan(ctx, cursor, match, count).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, tempKeys...)
		if cursor == 0 {
			break
		}
	}
	return keys, err
}

// DelAll atomically deletes all keys from given pattern, using KEYS before DEL.
func DelAll(ctx context.Context, client *redis.Client, pattern string) (tot int64, err error) {
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}

	return client.Del(ctx, keys...).Result()
}

// DelAllByScan atomically deletes all keys from given pattern, using SCAN before DEL.
func DelAllByScan(ctx context.Context, client *redis.Client, pattern string, scanCount int64) (tot int64, err error) {
	keys, err := ScanAll(ctx, client, pattern, scanCount)
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}

	return client.Del(ctx, keys...).Result()
}
