package xredis

import (
	"context"
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"testing"
	"time"
)

const (
	redisAddr       = "localhost:6379"
	redisPasswd     = "123"
	redisFakePasswd = "1234"
	redisDB         = 0
)

func TestHelper(t *testing.T) {
	for _, tc := range []struct {
		giveAddr   string
		givePasswd string
		wantErr    bool
	}{
		{redisAddr, redisPasswd, false},
		{redisAddr, redisFakePasswd, true},
	} {
		client := redis.NewClient(&redis.Options{
			Addr:     tc.giveAddr,
			Password: tc.givePasswd,
			DB:       redisDB,
		})
		l := logrus.New()
		l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
		redis.SetLogger(NewSilenceLogger())
		client.AddHook(NewLogrusLogger(l))

		t.Run("DeleteAll", func(t *testing.T) {
			client.Set(context.Background(), "test_a", "test_aaa", 0)
			client.Set(context.Background(), "test_b", "test_bbb", 0)
			client.Set(context.Background(), "test_c", "test_ccc", 0)

			tot, del, err := DelAll(client, "test_")
			if tc.wantErr {
				xtesting.NotNil(t, err)
			} else {
				xtesting.Equal(t, tot, 0)
				xtesting.Equal(t, del, 0)
				xtesting.Nil(t, err)
			}

			tot, del, err = DelAll(client, "test_a")
			if tc.wantErr {
				xtesting.NotNil(t, err)
			} else {
				xtesting.Equal(t, tot, 1)
				xtesting.Equal(t, del, 1)
				xtesting.Nil(t, err)
			}

			tot, del, err = DelAll(client, "test_*")
			if tc.wantErr {
				xtesting.NotNil(t, err)
			} else {
				xtesting.Equal(t, tot, 2)
				xtesting.Equal(t, del, 2)
				xtesting.Nil(t, err)
			}
		})

		t.Run("SetAll", func(t *testing.T) {
			for _, ttc := range []struct {
				giveKeys   []string
				giveValues []string
				wantOk     bool
			}{
				{[]string{}, []string{}, true},
				{[]string{"k"}, []string{}, false},
				{[]string{}, []string{"v"}, false},
				{[]string{"test_a", "test_b", "test_c"}, []string{"test_aaa", "test_bbb", "test_ccc"}, !tc.wantErr /* && true */},
			} {
				tot, add, err := SetAll(client, ttc.giveKeys, ttc.giveValues)
				if !ttc.wantOk {
					xtesting.NotNil(t, err)
				} else {
					xtesting.Nil(t, err)
					xtesting.Equal(t, tot, len(ttc.giveKeys))
					xtesting.Equal(t, add, len(ttc.giveKeys))
					for idx := range ttc.giveKeys {
						k, v := ttc.giveKeys[idx], ttc.giveValues[idx]
						xtesting.Equal(t, client.Get(context.Background(), k).Val(), v)
					}
				}
			}
		})

		t.Run("SetExAll", func(t *testing.T) {
			for _, ttc := range []struct {
				giveKeys   []string
				giveValues []string
				giveExs    []int64
				wantOk     bool
			}{
				{[]string{}, []string{}, []int64{}, true},
				{[]string{"k"}, []string{}, []int64{0}, false},
				{[]string{}, []string{"v"}, []int64{0}, false},
				{[]string{"k"}, []string{"v"}, []int64{}, false},
				{[]string{"test_a", "test_b", "test_c"}, []string{"test_aaa", "test_bbb", "test_ccc"}, []int64{1, 1, 1}, !tc.wantErr /* && true */},
			} {
				tot, add, err := SetExAll(client, ttc.giveKeys, ttc.giveValues, ttc.giveExs)
				if !ttc.wantOk {
					xtesting.NotNil(t, err)
				} else {
					xtesting.Nil(t, err)
					xtesting.Equal(t, tot, len(ttc.giveKeys))
					xtesting.Equal(t, add, len(ttc.giveKeys))
					for idx := range ttc.giveKeys {
						k, v := ttc.giveKeys[idx], ttc.giveValues[idx]
						xtesting.Equal(t, client.Get(context.Background(), k).Val(), v)
					}

					maxWait := int64(-1)
					for _, s := range ttc.giveExs {
						if s > maxWait {
							maxWait = s
						}
					}
					time.Sleep(time.Second * time.Duration(maxWait+1))
					for idx := range ttc.giveKeys {
						xtesting.NotNil(t, client.Get(context.Background(), ttc.giveKeys[idx]).Err())
					}
				}
			}
		})
	}
}

func TestLogger(t *testing.T) {
	l1 := logrus.New()
	l1.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	l2 := log.New(os.Stderr, "", log.LstdFlags)

	for _, tc := range []struct {
		name   string
		logger redis.Hook
	}{
		{"default", nil},
		{"logrus", NewLogrusLogger(l1)},
		{"logrus_no_err", NewLogrusLogger(l1, WithLogErr(false))},
		{"logger", NewLoggerLogger(l2)},
		{"logger_no_err", NewLoggerLogger(l2, WithLogErr(false))},
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
			}

			client.Get(context.Background(), "test" )     // String err
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

			client.Scan(context.Background(), 0, "test", 10)                             // Scan
			defer client.Del(context.Background(), "myhash")
			client.HSet(context.Background(), "myhash", "1", "111")                      // IntCmd
			client.HSet(context.Background(), "myhash", "2", "222")                      // IntCmd
			client.HGet(context.Background(), "myhash", "1")                             // StringCmd
			client.HGetAll(context.Background(), "myhash")                               // StringStringMapCmd
			defer client.Del(context.Background(), "myset")
			client.SAdd(context.Background(), "myset", "1", "2", "3")                    // IntCmd
			client.SMembers(context.Background(), "myset")                               // StringSliceCmd
			client.SPop(context.Background(), "myset")                                   // StringCmd
			client.SMembersMap(context.Background(), "myset")                            // StringStructMapCmd
			defer client.Del(context.Background(), "myzset")
			client.ZAdd(context.Background(), "myzset", &redis.Z{Score: 1, Member: "A"}) // IntCmd
			client.ZAdd(context.Background(), "myzset", &redis.Z{Score: 2, Member: "B"}) // IntCmd
			client.ZRange(context.Background(), "myzset", 0, 2)                          // StringSliceCmd
			client.ZRangeWithScores(context.Background(), "myzset", 0, 2)                // StringSliceCmd

			client.BitField(context.Background(), "BitField", "INCRBY", "i5", 100, 1, "GET", "u4", 0) // IntSliceCmd
			client.ScriptExists(context.Background(), "ScriptExists")                                 // BoolSliceCmd
			client.SlowLogGet(context.Background(), 1)                                                // SlowLogCmd
			client.Command(context.Background())                                                      // CommandsInfoCmd
		})
	}
}
