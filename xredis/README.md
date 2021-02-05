# xredis

## Dependencies

+ github.com/go-redis/redis
+ github.com/sirupsen/logrus

## Documents

### Types

+ `type ILogger interface`
+ `type SilenceLogger struct`
+ `type LogrusLogger struct`
+ `type LoggerLogger struct`

### Variables

+ None

### Constants

+ None

### Functions

+ `func DelAll(client *redis.Client, pattern string) (tot, del int, err error)`
+ `func SetAll(client *redis.Client, keys, values []string) (tot, add int, err error)`
+ `func SetExAll(client *redis.Client, keys, values []string, expirations []int64) (tot, add int, err error)`
+ `func NewSilenceLogger() *SilenceLogger`
+ `func NewLogrusLogger(logger *logrus.Logger) *LogrusLogger`
+ `func NewLoggerLogger(logger logrus.StdLogger) *LoggerLogger`

### Methods

+ `func (l *SilenceLogger) Printf(context.Context, string, ...interface{})`
+ `func (l *LogrusLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error`
+ `func (l *LoggerLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error`
