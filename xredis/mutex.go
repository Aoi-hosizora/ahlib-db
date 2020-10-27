package xredis

import (
	"github.com/gomodule/redigo/redis"
	"sync"
)

// MutexRedis will add a mutex lock to redis.Conn, to solve the concurrent problems:
// 	short write
// 	use of closed network connection
type MutexRedis struct {
	redis.Conn
	mu sync.Mutex
}

// NewMutexRedis creates a new MutexRedis.
func NewMutexRedis(conn redis.Conn) *MutexRedis {
	return &MutexRedis{Conn: conn}
}

func (m *MutexRedis) Do(commandName string, args ...interface{}) (interface{}, error) {
	m.mu.Lock()
	reply, err := m.Conn.Do(commandName, args...)
	m.mu.Unlock()
	return reply, err
}
