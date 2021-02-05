package xredis

import (
	"context"
	"github.com/Aoi-hosizora/ahlib/xtesting"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"testing"
	"time"
)

const (
	redisAddr   = "localhost:6379"
	redisPasswd = "123"
	redisDB     = 0
)

func TestHelper(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPasswd,
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
		xtesting.Equal(t, tot, 0)
		xtesting.Equal(t, del, 0)
		xtesting.Nil(t, err)
		tot, del, err = DelAll(client, "test_a")
		xtesting.Equal(t, tot, 1)
		xtesting.Equal(t, del, 1)
		xtesting.Nil(t, err)
		tot, del, err = DelAll(client, "test_*")
		xtesting.Equal(t, tot, 2)
		xtesting.Equal(t, del, 2)
		xtesting.Nil(t, err)
	})

	t.Run("SetAll", func(t *testing.T) {
		for _, tc := range []struct {
			giveKeys   []string
			giveValues []string
			wantOk     bool
		}{
			{[]string{}, []string{}, true},
			{[]string{"k"}, []string{}, false},
			{[]string{}, []string{"v"}, false},
			{[]string{"test_a", "test_b", "test_c"}, []string{"test_aaa", "test_bbb", "test_ccc"}, true},
		} {
			tot, add, err := SetAll(client, tc.giveKeys, tc.giveValues)
			if !tc.wantOk {
				xtesting.NotNil(t, err)
			} else {
				xtesting.Nil(t, err)
				xtesting.Equal(t, tot, len(tc.giveKeys))
				xtesting.Equal(t, add, len(tc.giveKeys))
				for idx := range tc.giveKeys {
					k, v := tc.giveKeys[idx], tc.giveValues[idx]
					xtesting.Equal(t, client.Get(context.Background(), k).Val(), v)
				}
			}
		}
	})

	t.Run("SetExAll", func(t *testing.T) {
		for _, tc := range []struct {
			giveKeys   []string
			giveValues []string
			giveExs    []int64
			wantOk     bool
		}{
			{[]string{}, []string{}, []int64{}, true},
			{[]string{"k"}, []string{}, []int64{0}, false},
			{[]string{}, []string{"v"}, []int64{0}, false},
			{[]string{"k"}, []string{"v"}, []int64{}, false},
			{[]string{"test_a", "test_b", "test_c"}, []string{"test_aaa", "test_bbb", "test_ccc"}, []int64{1, 1, 1}, true},
		} {
			tot, add, err := SetExAll(client, tc.giveKeys, tc.giveValues, tc.giveExs)
			if !tc.wantOk {
				xtesting.NotNil(t, err)
			} else {
				xtesting.Nil(t, err)
				xtesting.Equal(t, tot, len(tc.giveKeys))
				xtesting.Equal(t, add, len(tc.giveKeys))
				for idx := range tc.giveKeys {
					k, v := tc.giveKeys[idx], tc.giveValues[idx]
					xtesting.Equal(t, client.Get(context.Background(), k).Val(), v)
				}

				maxWait := int64(-1)
				for _, s := range tc.giveExs {
					if s > maxWait {
						maxWait = s
					}
				}
				time.Sleep(time.Second * time.Duration(maxWait+1))
				for idx := range tc.giveKeys {
					xtesting.NotNil(t, client.Get(context.Background(), tc.giveKeys[idx]).Err())
				}
			}
		}
	})
}

func TestLogger(t *testing.T) {

}
