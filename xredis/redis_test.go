package xredis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"log"
	"testing"
	"time"
)

/*

func TestLogrus(t *testing.T) {
	conn, err := redis.Dial("tcp", "localhost:6379", redis.DialPassword("123"), redis.DialDatabase(1))
	if err != nil {
		log.Fatalln(err)
	}

	conn = NewLogrusRedis(conn, logrus.New(), true).WithSkip(3)
	conn = NewMutexRedis(conn)

	_, _ = conn.Do("GET", "aaaaa-a")
	_, _ = conn.Do("SET", "aaaaa-a", "abc")
	_, _ = conn.Do("SET", "aaaaa-b", "bcd")
	_, _ = conn.Do("GET", "aaaaa-a")
	_, _ = conn.Do("KEYS", "aaaaa-*")
	_, _, _ = WithConn(conn).DeleteAll("aaaaa-*")
}

func TestLogger(t *testing.T) {
	conn, err := redis.Dial("tcp", "localhost:6379", redis.DialPassword("123"), redis.DialDatabase(1))
	if err != nil {
		log.Fatalln(err)
	}

	conn = NewLoggerRedis(conn, log.New(os.Stderr, "", log.LstdFlags), true).WithSkip(3)
	conn = NewMutexRedis(conn)

	_, _ = conn.Do("GET", "aaaaa-a")
	_, _ = conn.Do("SET", "aaaaa-a", "abc")
	_, _ = conn.Do("SET", "aaaaa-b", "bcd")
	_, _ = conn.Do("GET", "aaaaa-a")
	_, _ = conn.Do("KEYS", "aaaaa-*")
	_, _, _ = WithConn(conn).DeleteAll("aaaaa-*")
}

func TestMutex(t *testing.T) {
	conn, err := redis.Dial("tcp", "localhost:6379", redis.DialPassword("123"), redis.DialDatabase(1))
	if err != nil {
		log.Fatalln(err)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{})
	conn = NewLogrusRedis(conn, logger, true).WithSkip(3)
	conn = NewMutexRedis(conn)

	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = conn.Do("GET", "aaaaa-a")
			wg.Done()
		}()
	}
	wg.Wait()
}

*/

func TestXXX(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123",
		DB:       1,
	})
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC3339})
	client.AddHook(NewLogrusLogger(l))
	log.Println(client.Set(context.Background(), "a", "test", 0).Result())
	log.Println(client.Get(context.Background(), "a").Result())
	log.Println(client.Scan(context.Background(), 0, "", 10).Result())
}
