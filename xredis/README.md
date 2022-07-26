# xredis

## Dependencies

+ github.com/Aoi-hosizora/ahlib
+ github.com/go-redis/redis/v8
+ github.com/sirupsen/logrus

## Documents

### Types

+ `type LoggerOption func`
+ `type ILogger interface`
+ `type SilenceLogger struct`
+ `type LogrusLogger struct`
+ `type StdLogger struct`
+ `type LoggerParam struct`

### Variables

+ `var FormatLoggerFunc func`
+ `var FieldifyLoggerFunc func`

### Constants

+ None

### Functions

+ `func ScanAll(ctx context.Context, client *redis.Client, match string, count int64) (keys []string, err error)`
+ `func ScanAllWithCallback(ctx context.Context, client *redis.Client, match string, count int64, callback func(keys []string)) error`
+ `func DelAll(ctx context.Context, client *redis.Client, pattern string) (tot int64, err error)`
+ `func DelAllByScan(ctx context.Context, client *redis.Client, pattern string, scanCount int64) (tot int64, err error)`
+ `func DelAllByScanCallback(ctx context.Context, client *redis.Client, pattern string, scanCount int64, ignoreDelError bool) (tot int64, err error)`
+ `func WithLogErr(log bool) LoggerOption`
+ `func WithLogCmd(log bool) LoggerOption`
+ `func WithSkip(skip int) LoggerOption`
+ `func WithSlowThreshold(threshold time.Duration) LoggerOption`
+ `func EnableLogger()`
+ `func DisableLogger()`
+ `func NewSilenceLogger() *SilenceLogger`
+ `func NewLogrusLogger(logger *logrus.Logger, options ...LoggerOption) *LogrusLogger`
+ `func NewStdLogger(logger logrus.StdLogger, options ...LoggerOption) *StdLogger`

### Methods

+ `func (s *SilenceLogger) Printf(context.Context, string, ...interface{})`
+ `func (l *LogrusLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error)`
+ `func (l *LogrusLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error`
+ `func (l *StdLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error)`
+ `func (l *StdLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error`
