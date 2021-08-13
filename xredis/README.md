# xredis

## Dependencies

+ github.com/go-redis/redis
+ github.com/sirupsen/logrus

## Documents

### Types

+ `type ILogger interface`
+ `type LoggerOption func`
+ `type SilenceLogger struct`
+ `type LogrusLogger struct`
+ `type LoggerLogger struct`

### Variables

+ None

### Constants

+ None

### Functions

+ `func ScanAll(ctx context.Context, client *redis.Client, match string, count int64) (keys []string, err error)`
+ `func DelAll(ctx context.Context, client *redis.Client, pattern string) (tot int64, err error)`
+ `func DelAllByScan(ctx context.Context, client *redis.Client, pattern string, scanCount int64) (tot int64, err error)`
+ `func WithLogErr(logErr bool) LoggerOption`
+ `func EnableLogger()`
+ `func DisableLogger()`
+ `func NewSilenceLogger() *SilenceLogger`
+ `func NewLogrusLogger(logger *logrus.Logger) *LogrusLogger`
+ `func NewLoggerLogger(logger logrus.StdLogger) *LoggerLogger`

### Methods

+ `func (l *SilenceLogger) Printf(context.Context, string, ...interface{})`
+ `func (l *LogrusLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error`
+ `func (l *LoggerLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error`
