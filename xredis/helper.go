package xredis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
)

// DelAll deletes all keys from given pattern (KEYS -> DEL). This is an atomic operator and it will return err when failed.
func DelAll(client *redis.Client, ctx context.Context, pattern string) (tot int64, err error) {
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}
	if len(keys) == 0 {
		return 0, nil
	}

	return client.Del(ctx, keys...).Result()
}

var (
	errDifferentKeyValueLength = errors.New("xredis: different length of keys and values")
)

// SetAll sets all given key-value pairs (MSET). This is an atomic operator and it will return err when failed.
func SetAll(client *redis.Client, ctx context.Context, keys, values []string) (tot int64, err error) {
	l := len(keys)
	if l != len(values) {
		return 0, errDifferentKeyValueLength
	}
	if l == 0 {
		return 0, nil
	}

	parameters := make([]interface{}, 0, l*2)
	for idx, key := range keys {
		parameters = append(parameters, key, values[idx])
	}
	err = client.MSet(ctx, parameters...).Err()
	if err != nil {
		return 0, err
	}
	return int64(l), nil
}
