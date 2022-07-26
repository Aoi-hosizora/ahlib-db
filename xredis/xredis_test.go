package xredis

import (
	"context"
	"errors"
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xerror"
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"log"
	"testing"
	"time"
)

const (
	redisAddr       = "localhost:6379"
	redisPasswd     = "123"
	redisFakePasswd = "1234"
	redisDB         = 0
)

func prepareTestABC(client *redis.Client) (delFunc func()) {
	client.Set(context.Background(), "test_a", "test_aaa", 0)
	client.Set(context.Background(), "test_b", "test_bbb", 0)
	client.Set(context.Background(), "test_c", "test_ccc", 0)
	return func() {
		client.Del(context.Background(), "test_a", "test_b", "test_c")
	}
}

type delBanner struct{}

func (delBanner) BeforeProcess(ctx context.Context, cmder redis.Cmder) (context.Context, error) {
	if cmder.Name() == "del" {
		return ctx, errors.New("del is forbidden to execute")
	}
	return ctx, nil
}

func (delBanner) AfterProcess(context.Context, redis.Cmder) error {
	return nil
}

func (delBanner) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (delBanner) AfterProcessPipeline(context.Context, []redis.Cmder) error {
	return nil
}

func TestHelper(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: redisAddr, Password: redisPasswd, DB: redisDB})
	redis.SetLogger(NewSilenceLogger()) // <<<
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	client.AddHook(NewLogrusLogger(l))
	defer client.Close()

	t.Run("ScanAll", func(t *testing.T) {
		delFunc := prepareTestABC(client)
		defer delFunc()

		keys, err := ScanAll(context.Background(), client, "test_*", 1)
		xtesting.Nil(t, err)
		xtesting.ElementMatch(t, keys, []string{"test_a", "test_b", "test_c"})
	})

	t.Run("ScanAllWithCallback", func(t *testing.T) {
		delFunc := prepareTestABC(client)
		defer delFunc()

		count := 0
		err := ScanAllWithCallback(context.Background(), client, "test_*", 1, func(keys []string, cursor uint64) (toContinue bool) {
			count++
			return false
		})
		xtesting.Equal(t, count, 1)
		xtesting.Nil(t, err)

		gotKeys := make([]string, 0)
		count = 0
		err = ScanAllWithCallback(context.Background(), client, "test_*", 1, func(keys []string, cursor uint64) (toContinue bool) {
			gotKeys = append(gotKeys, keys...)
			count++
			return true
		})
		xtesting.Nil(t, err)
		xtesting.ElementMatch(t, gotKeys, []string{"test_a", "test_b", "test_c"})
		xtesting.Equal(t, count >= 3, true)
	})

	t.Run("DelAll", func(t *testing.T) {
		prepareTestABC(client)

		tot, err := DelAll(context.Background(), client, "test_")
		xtesting.Nil(t, err)
		xtesting.Equal(t, tot, int64(0))
		tot, err = DelAll(context.Background(), client, "test_a")
		xtesting.Nil(t, err)
		xtesting.Equal(t, tot, int64(1))
		tot, err = DelAll(context.Background(), client, "test_*")
		xtesting.Nil(t, err)
		xtesting.Equal(t, tot, int64(2))
		keys, err := client.Keys(context.Background(), "test_*").Result()
		xtesting.Nil(t, err)
		xtesting.EmptyCollection(t, keys)
	})

	t.Run("DelAllByScan", func(t *testing.T) {
		prepareTestABC(client)

		tot, err := DelAllByScan(context.Background(), client, "test_", 1)
		xtesting.Nil(t, err)
		xtesting.Equal(t, tot, int64(0))
		tot, err = DelAllByScan(context.Background(), client, "test_a", 1)
		xtesting.Nil(t, err)
		xtesting.Equal(t, tot, int64(1))
		tot, err = DelAllByScan(context.Background(), client, "test_*", 1)
		xtesting.Nil(t, err)
		xtesting.Equal(t, tot, int64(2))
		keys, err := client.Keys(context.Background(), "test_*").Result()
		xtesting.Nil(t, err)
		xtesting.EmptyCollection(t, keys)
	})

	t.Run("DelAllByScanCallback", func(t *testing.T) {
		prepareTestABC(client)

		tot, err := DelAllByScanCallback(context.Background(), client, "test_", 1, true)
		xtesting.Nil(t, err)
		xtesting.Equal(t, tot, int64(0))
		tot, err = DelAllByScanCallback(context.Background(), client, "test_a", 1, true)
		xtesting.Nil(t, err)
		xtesting.Equal(t, tot, int64(1))
		tot, err = DelAllByScanCallback(context.Background(), client, "test_*", 1, true)
		xtesting.Nil(t, err)
		xtesting.Equal(t, tot, int64(2))
		keys, err := client.Keys(context.Background(), "test_*").Result()
		xtesting.Nil(t, err)
		xtesting.EmptyCollection(t, keys)
	})

	t.Run("Errors", func(t *testing.T) {
		c1 := redis.NewClient(&redis.Options{Addr: redisAddr, Password: redisFakePasswd, DB: redisDB})
		c1.AddHook(NewLogrusLogger(l))
		defer c1.Close()
		c2 := redis.NewClient(&redis.Options{Addr: redisAddr, Password: redisPasswd, DB: redisDB})
		c2.AddHook(NewLogrusLogger(l))
		defer c2.Close()

		// ScanAll
		_, err := ScanAll(context.Background(), c1, "test_*", 1)
		xtesting.NotNil(t, err)

		// ScanAllWithCallback
		err = ScanAllWithCallback(context.Background(), c1, "test_*", 1, func(keys []string, cursor uint64) (toContinue bool) { return true })
		xtesting.NotNil(t, err)
		xtesting.Panic(t, func() { _ = ScanAllWithCallback(context.Background(), c1, "test_", 1, nil) })

		// DelAll
		_, err = DelAll(context.Background(), c1, "test_*")
		xtesting.NotNil(t, err)

		// DelAllByScan
		_, err = DelAllByScan(context.Background(), c1, "test_*", 1)
		xtesting.NotNil(t, err)

		// DelAllByScanCallback
		_, err = DelAllByScanCallback(context.Background(), c1, "test_*", 1, true)
		xtesting.NotNil(t, err)

		// DelAllByScanCallback with delBanner
		c2.AddHook(delBanner{})
		delFunc := prepareTestABC(client)
		defer delFunc()
		tot, err := DelAllByScanCallback(context.Background(), c2, "test_*", 1, false)
		xtesting.Equal(t, len(xerror.Separate(err)), 1)
		xtesting.Equal(t, tot, 0)
		tot, err = DelAllByScanCallback(context.Background(), c2, "test_*", 1, true)
		xtesting.Equal(t, len(xerror.Separate(err)) >= 3, true)
		xtesting.Equal(t, tot, 0)
	})
}

func TestLogger(t *testing.T) {
	l1 := logrus.New()
	l1.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	l2 := log.Default()

	for _, tc := range []struct {
		name   string
		enable bool
		custom bool
		logger redis.Hook
	}{
		{"default", true, false, nil},
		//
		{"logrus", true, false, NewLogrusLogger(l1, WithSlowThreshold(time.Millisecond*1))},
		{"logrus_custom", true, true, NewLogrusLogger(l1)},
		{"logrus_no_err", true, false, NewLogrusLogger(l1, WithLogErr(false))},
		{"logrus_no_cmd", true, false, NewLogrusLogger(l1, WithLogCmd(false))},
		{"logrus_disable", false, false, NewLogrusLogger(l1)},
		//
		{"stdlog", true, false, NewStdLogger(l2, WithSkip(4), WithSlowThreshold(time.Millisecond*1))},
		{"stdlog_custom", true, true, NewStdLogger(l2, WithSkip(4))},
		{"stdlog_no_xxx", true, false, NewStdLogger(l2, WithLogErr(false), WithLogCmd(false))},
		{"stdlog_disable", false, false, NewStdLogger(l2)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := redis.NewClient(&redis.Options{
				Addr:     redisAddr,
				Password: redisPasswd,
				DB:       0,
			})
			redis.SetLogger(NewSilenceLogger())
			if tc.logger != nil {
				client.AddHook(tc.logger)
				_, _ = tc.logger.BeforeProcessPipeline(context.Background(), nil)
				_ = tc.logger.AfterProcessPipeline(context.Background(), nil)
				_ = tc.logger.AfterProcess(context.Background(), nil)
				NewSilenceLogger().Printf(context.Background(), "")
			}
			if tc.enable {
				EnableLogger()
			} else {
				DisableLogger()
			}
			if tc.custom {
				FormatLoggerFunc = func(p *LoggerParam) string {
					if p.ErrorMsg != "" {
						return fmt.Sprintf("[Redis] err: %v - %s - %s", p.ErrorMsg, p.Command, p.Source)
					}
					return fmt.Sprintf("[Redis] %6s - %12s - %s - %s", p.Status, p.Duration.String(), p.Command, p.Source)
				}
				FieldifyLoggerFunc = func(p *LoggerParam) logrus.Fields {
					return logrus.Fields{"module": "redis"}
				}
				defer func() {
					FormatLoggerFunc = nil
					FieldifyLoggerFunc = nil
				}()
			}

			client.Do(context.Background(), "XXX")
			client.Do(context.Background(), "SET", "X", "X", "X", "X")
			client.Get(context.Background(), "test")               // String err
			client.Set(context.Background(), "test", "test", 0)    // Status
			client.Get(context.Background(), "test")               // String
			client.Exists(context.Background(), "test", "xxx")     // Bool
			client.Set(context.Background(), "test1", "test 1", 0) // Status
			client.Set(context.Background(), "test2", "test 2", 0) // Status
			defer client.Del(context.Background(), "test")
			defer client.Del(context.Background(), "test1")
			defer client.Del(context.Background(), "test2")
			client.Get(context.Background(), "test")                    // String
			client.MGet(context.Background(), "test", "test1", "test2") // SliceCmd
			client.Keys(context.Background(), "tes*")                   // StringSlice
			client.Del(context.Background(), "test", "test1", "test2")  // Int
			client.Set(context.Background(), "F", 1.1, 0)               // Status
			client.Set(context.Background(), "I", 1, 0)                 // Status
			client.Incr(context.Background(), "I")                      // Int
			client.IncrByFloat(context.Background(), "F", 1)            // Float
			client.Del(context.Background(), "F")                       // Int
			client.Del(context.Background(), "I")                       // Int

			client.Scan(context.Background(), 0, "test", 10) // Scan
			defer client.Del(context.Background(), "myhash")
			client.HSet(context.Background(), "myhash", "1", "111") // IntCmd
			client.HSet(context.Background(), "myhash", "2", "222") // IntCmd
			client.HGet(context.Background(), "myhash", "1")        // StringCmd
			client.HGetAll(context.Background(), "myhash")          // StringStringMapCmd
			client.HExists(context.Background(), "myhash", "1")     // BoolCmd
			client.HExists(context.Background(), "myhash", "0")     // BoolCmd
			defer client.Del(context.Background(), "myset")
			client.SAdd(context.Background(), "myset", "1", "2", "3") // IntCmd
			client.SMembers(context.Background(), "myset")            // StringSliceCmd
			client.SPop(context.Background(), "myset")                // StringCmd
			client.SMembersMap(context.Background(), "myset")         // StringStructMapCmd
			defer client.Del(context.Background(), "myzset")
			client.ZAdd(context.Background(), "myzset", &redis.Z{Score: 1, Member: "A"}) // IntCmd
			client.ZAdd(context.Background(), "myzset", &redis.Z{Score: 2, Member: "B"}) // IntCmd
			client.ZRange(context.Background(), "myzset", 0, 2)                          // StringSliceCmd
			client.ZRangeWithScores(context.Background(), "myzset", 0, 2)                // StringSliceCmd

			client.BitField(context.Background(), "BitField", "INCRBY", "i5", 100, 1, "GET", "u4", 0) // IntSliceCmd
			client.ScriptExists(context.Background(), "ScriptExists")                                 // BoolSliceCmd
			client.SlowLogGet(context.Background(), 1)                                                // SlowLogCmd
			client.Command(context.Background())                                                      // CommandsInfoCmd
			client.PubSubNumSub(context.Background(), "")                                             // StringIntMapCmd
			client.TTL(context.Background(), "test")                                                  // DurationCmd
			client.Time(context.Background())                                                         // TimeCmd
		})
	}
}
