# xredis

## Dependencies

+ github.com/Aoi-hosizora/ahlib
+ github.com/go-redis/redis
+ github.com/sirupsen/logrus

## Documents

### Types

+ `type LoggerOption func`
+ `type ILogger interface`
+ `type SilenceLogger struct`
+ `type LogrusLogger struct`
+ `type StdLogger struct`

### Variables

+ None

### Constants

+ None

### Functions

+ `func ScanAll(ctx context.Context, client *redis.Client, match string, count int64) (keys []string, err error)`
+ `func DelAll(ctx context.Context, client *redis.Client, pattern string) (tot int64, err error)`
+ `func DelAllByScan(ctx context.Context, client *redis.Client, pattern string, scanCount int64) (tot int64, err error)`
+ `func WithLogErr(log bool) LoggerOption`
+ `func WithLogCmd(log bool) LoggerOption`
+ `func WithSkip(skip int) LoggerOption`
+ `func EnableLogger()`
+ `func DisableLogger()`
+ `func NewSilenceLogger() *SilenceLogger`
+ `func NewLogrusLogger(logger *logrus.Logger, options ...LoggerOption) *LogrusLogger`
+ `func NewStdLogger(logger logrus.StdLogger, options ...LoggerOption) *StdLogger`

### Methods

+ `func (s *SilenceLogger) Printf(context.Context, string, ...interface{})`
+ `func (l *LogrusLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error)`
+ `func (l *LogrusLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error`
+ `func (s *StdLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error)`
+ `func (s *StdLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error`
