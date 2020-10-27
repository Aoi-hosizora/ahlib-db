package xredis

import (
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"sync"
	"testing"
)

func TestLogrus(t *testing.T) {
	conn, err := redis.Dial("tcp", "localhost:6379", redis.DialPassword("123"), redis.DialDatabase(1))
	if err != nil {
		log.Fatalln(err)
	}

	conn = NewLogrusLogger(conn, logrus.New(), true).WithSkip(3)
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
	conn = NewLogrusLogger(conn, logger, true).WithSkip(3)
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
