package xredis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
)

// DelAll deletes all keys from given pattern (KEYS -> DEL). This is an atomic operator and return err when failed.
func DelAll(client *redis.Client, pattern string) (tot, del int, err error) {
	keys, err := client.Keys(context.Background(), pattern).Result()
	if err != nil {
		return 0, 0, err
	}
	tot = len(keys)
	if tot == 0 {
		return 0, 0, nil
	}

	cnt, err := client.Del(context.Background(), keys...).Result()
	if err != nil {
		return 0, 0, err
	}
	return tot, int(cnt), nil
}

var (
	errDifferentKeyValueLength    = errors.New("xredis: different length of keys and values")
	errDifferentKeyValueExpLength = errors.New("xredis: different length of keys, values and expirations")
)

// SetAll sets all given key-value pairs (SET -> SET -> ...). This is a non-atomic operator, that means if there is a value failed to set,
// no rollback will be done, and it will return the current added count and error value.
func SetAll(client *redis.Client, keys, values []string) (tot, add int, err error) {
	tot = len(keys)
	if tot != len(values) {
		return 0, 0, errDifferentKeyValueLength
	}

	var someErr error
	for idx, key := range keys {
		val := values[idx]
		e := client.Set(context.Background(), key, val, 0).Err()
		if e == nil {
			add++
		} else if someErr == nil {
			someErr = e
		}
	}

	return tot, add, someErr
}

// SetExAll sets all given key-value-expiration pairs (SET -> SET -> ...), equals to SetAll with expiration in second. This is a non-atomic operator,
// that means if there is a value failed to set, no rollback will be done, and it will return the current added count and error value.
func SetExAll(client *redis.Client, keys, values []string, expirations []int64) (tot, add int, err error) {
	tot = len(keys)
	if tot != len(values) || tot != len(expirations) {
		return 0, 0, errDifferentKeyValueExpLength
	}

	var someErr error
	for idx, key := range keys {
		val := values[idx]
		ex := expirations[idx]
		e := client.Set(context.Background(), key, val, time.Duration(ex*1e9)).Err()
		if e == nil {
			add++
		} else if someErr == nil {
			someErr = e
		}
	}

	return tot, add, someErr
}
