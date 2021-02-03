# xredis

## Dependencies

+ github.com/go-redis/redis

## Documents

### Types

+ None

### Variables

+ None

### Constants

+ None

### Functions

+ `func DelAll(client *redis.Client, pattern string) (tot, del int, err error)`
+ `func SetAll(client *redis.Client, keys, values []string) (tot, add int, err error)`
+ `func SetExAll(client *redis.Client, keys, values []string, expirations []int64) (tot, add int, err error)`

### Methods

+ None

---

+ `type LogrusRedis struct {}`
+ `NewLogrusRedis(conn redis.Conn, logger *logrus.Logger, logMode bool) *LogrusRedis`
+ `type LoggerRedis struct {}`
+ `NewLoggerRedis(conn redis.Conn, logger *log.Logger, logMode bool) *LoggerRedis`
+ `type MutexRedis struct {}`
+ `NewMutexRedis(conn redis.Conn) *MutexRedis`
