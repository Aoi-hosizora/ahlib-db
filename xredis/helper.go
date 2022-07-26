package xredis

import (
	"context"
	"github.com/Aoi-hosizora/ahlib/xerror"
	"github.com/go-redis/redis/v8"
)

// ScanAll scans all keys from given pattern and scan count, and returns all keys when finish scanning, is a wrapper function for SCAN command.
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
	return keys, nil
}

const (
	panicNilScanAllCallback = "xredis: nil scan all callback"
)

// ScanAllWithCallback scans all keys from given pattern and scan count, and invokes callback when get some keys, is a wrapper function for SCAN command.
func ScanAllWithCallback(ctx context.Context, client *redis.Client, match string, count int64, callback func(keys []string, cursor uint64) (toContinue bool)) error {
	if callback == nil {
		panic(panicNilScanAllCallback)
	}
	cursor := uint64(0)
	var tempKeys []string
	var err error
	for {
		tempKeys, cursor, err = client.Scan(ctx, cursor, match, count).Result()
		if err != nil {
			return err
		}
		if !callback(tempKeys, cursor) {
			break
		}
		if cursor == 0 {
			break
		}
	}
	return nil
}

// DelAll atomically deletes all keys from given pattern, using KEYS command to get all keys before DEL command.
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

// DelAllByScan atomically deletes all keys from given pattern, using SCAN command to get all keys before DEL command.
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

// DelAllByScanCallback atomically deletes all keys from given pattern, using SCAN command with callback and DEL command multi times.
func DelAllByScanCallback(ctx context.Context, client *redis.Client, pattern string, scanCount int64, ignoreDelError bool) (tot int64, err error) {
	tot = 0
	errs := make([]error, 0)
	err = ScanAllWithCallback(ctx, client, pattern, scanCount, func(keys []string, cursor uint64) (toContinue bool) {
		count, delErr := client.Del(ctx, keys...).Result()
		if delErr != nil {
			errs = append(errs, delErr)
			if !ignoreDelError {
				return false // break scan
			}
		}
		tot += count
		return true // continue
	})

	if err != nil {
		errs = append(errs, err)
	}
	return tot, xerror.Combine(errs...) // tot may be still larger than zero even if errs is not empty
}
