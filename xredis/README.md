# xredis

### Functions

+ `type LogrusRedis struct {}`
+ `NewLogrusRedis(conn redis.Conn, logger *logrus.Logger, logMode bool) *LogrusRedis`
+ `type LoggerRedis struct {}`
+ `NewLoggerRedis(conn redis.Conn, logger *log.Logger, logMode bool) *LoggerRedis`
+ `type MutexRedis struct {}`
+ `NewMutexRedis(conn redis.Conn) *MutexRedis`

### Helper functions

+ `type Helper struct {}`
+ ` WithConn(conn redis.Conn) *Helper`
+ `(h *Helper) DeleteAll(pattern string) (total int, del int, err error)`
+ `(h *Helper) SetAll(keys []string, values []string) (total int, add int, err error)`
+ `(h *Helper) SetExAll(keys []string, values []string, exs []int64) (total int, add int, err error)`
