package xredis

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// loggerOptions is a type of LogrusLogger and StdLogger's option, each field can be set by LoggerOption function type.
type loggerOptions struct {
	logErr bool
	logCmd bool
	skip   int
}

// LoggerOption represents an option type for LogrusLogger's option and StdLogger's option, can be created by WithXXX functions.
type LoggerOption func(*loggerOptions)

// WithLogErr creates a LoggerOption to decide whether to do log for errors or not, defaults to true.
func WithLogErr(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logErr = log
	}
}

// WithLogCmd creates a LoggerOption to decide whether to do log for redis commands or not, defaults to true.
func WithLogCmd(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logCmd = log
	}
}

// WithSkip creates a LoggerOption to specific runtime skip for getting runtime information, defaults to 4.
func WithSkip(skip int) LoggerOption {
	return func(o *loggerOptions) {
		o.skip = skip
	}
}

var (
	// _enable is a global flag to control behaviors of LogrusLogger and StdLogger.
	_enable = true

	// _enableMu locks _enable.
	_enableMu sync.RWMutex
)

// EnableLogger enables LogrusLogger and StdLogger to do any log.
func EnableLogger() {
	_enableMu.Lock()
	_enable = true
	_enableMu.Unlock()
}

// DisableLogger disables LogrusLogger and StdLogger.
func DisableLogger() {
	_enableMu.Lock()
	_enable = false
	_enableMu.Unlock()
}

// ILogger abstracts redis's internal logger to an interface.
type ILogger interface {
	Printf(ctx context.Context, format string, v ...interface{})
}

// SilenceLogger represents a redis's logger, used to hide all logs, including error message.
type SilenceLogger struct{}

// NewSilenceLogger creates a new SilenceLogger.
//
// Example:
// 	client := redis.NewClient(options)
// 	redis.SetLogger(xredis.NewSilenceLogger())
func NewSilenceLogger() *SilenceLogger {
	return &SilenceLogger{}
}

// Printf implements redis.Logging interface, it does nothing for logging.
func (s *SilenceLogger) Printf(context.Context, string, ...interface{}) {}

// LogrusLogger represents a redis.Hook as redis's logger, used to log redis's executing message to logrus.Logger.
type LogrusLogger struct {
	redis.Hook
	logger *logrus.Logger
	option *loggerOptions
}

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger and LoggerOption-s.
//
// Example:
// 	client := redis.NewClient(...)
// 	redis.SetLogger(NewSilenceLogger()) // must silence fist
// 	l := logrus.New()
// 	l.SetFormatter(&logrus.TextFormatter{})
// 	client.AddHook(xredis.NewLogrusLogger(l))
func NewLogrusLogger(logger *logrus.Logger, options ...LoggerOption) *LogrusLogger {
	opt := &loggerOptions{logErr: true, logCmd: true, skip: 4}
	for _, o := range options {
		if o != nil {
			o(opt)
		}
	}
	return &LogrusLogger{logger: logger, option: opt}
}

// StdLogger represents a redis.Hook as redis's logger, used to log redis's executing message to logrus.StdLogger.
type StdLogger struct {
	redis.Hook
	logger logrus.StdLogger
	option *loggerOptions
}

// NewStdLogger creates a new StdLogger using given logrus.StdLogger and LoggerOption-s.
//
// Example:
// 	client := redis.NewClient(...)
// 	redis.SetLogger(NewSilenceLogger()) // must silence fist
// 	l := log.Default()
// 	client.AddHook(xredis.NewStdLogger(l))
func NewStdLogger(logger logrus.StdLogger, options ...LoggerOption) *StdLogger {
	opt := &loggerOptions{logErr: true, logCmd: true, skip: 4}
	for _, o := range options {
		if o != nil {
			o(opt)
		}
	}
	return &StdLogger{logger: logger, option: opt}
}

// =======
// methods
// =======

// &_startTimeKey is the key for storing processing start time to context.
var _startTimeKey int

// BeforeProcess implements redis.Hook interface, it saves start time to context, and will be used in AfterProcess.
func (l *LogrusLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, &_startTimeKey, time.Now()), nil
}

// BeforeProcess implements redis.Hook interface, it saves start time to context, and will be used in AfterProcess.
func (s *StdLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, &_startTimeKey, time.Now()), nil
}

// AfterProcess implements redis.Hook interface, it logs redis's message to logrus.Logger.
func (l *LogrusLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	endTime := time.Now()
	startTime, ok := ctx.Value(&_startTimeKey).(time.Time)
	_enableMu.RLock()
	enable := _enable
	_enableMu.RUnlock()
	if !enable || !ok {
		return nil // ignore
	}

	_, file, line, _ := runtime.Caller(l.option.skip) // defaults to 1
	source := fmt.Sprintf("%s:%d", file, line)
	m, f, isErr := extractLoggerData(cmd, endTime.Sub(startTime), source)
	if isErr && l.option.logErr {
		l.logger.WithFields(f).Error(m)
	} else if !isErr && l.option.logCmd {
		l.logger.WithFields(f).Info(m)
	}

	return nil
}

// AfterProcess implements redis.Hook interface, it logs redis's message to logrus.StdLogger.
func (s *StdLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	endTime := time.Now()
	startTime, ok := ctx.Value(&_startTimeKey).(time.Time)
	_enableMu.RLock()
	enable := _enable
	_enableMu.RUnlock()
	if !enable || !ok {
		return nil // ignore
	}

	_, file, line, _ := runtime.Caller(s.option.skip) // defaults to 1
	source := fmt.Sprintf("%s:%d", file, line)
	m, _, isErr := extractLoggerData(cmd, endTime.Sub(startTime), source)
	if (isErr && s.option.logErr) || (!isErr && s.option.logCmd) {
		s.logger.Print(m)
	}

	return nil
}

// ========
// internal
// ========

// extractLoggerData extracts and formats given redis.Cmder, time.Duration and source string to logger message and logrus.Fields.
//
// Logs like:
// 	[Redis] ERR invalid password | SET test_a test_aaa | F:/Projects/ahlib-db/xredis/redis_test.go:41
// 	[Redis]    Nil |   305.9909ms | GET test | F:/Projects/ahlib-db/xredis/redis_test.go:126
// 	[Redis]      3 |      995.5µs | KEYS tes* | F:/Projects/ahlib-db/xredis/redis_test.go:146
// 	[Redis]     2i |      997.8µs | DEL test_ test_c | F:/Projects/ahlib-db/xredis/helper.go:21
// 	[Redis]      F |      997.3µs | HEXISTS test xxx | F:/Projects/ahlib-db/xredis/redis_test.go:141
// 	[Redis]     OK |    25.9306ms | SET 'test num' 1 | F:/Projects/ahlib-db/xredis/redis_test.go:59
// 	       |------| |------------| |----------------| |--------------------------------------------|
// 	          6           12               ...                               ...
func extractLoggerData(cmd redis.Cmder, duration time.Duration, source string) (string, logrus.Fields, bool) {
	var msg string
	var fields logrus.Fields
	var isErr bool

	err := cmd.Err()
	isnil := errors.Is(err, redis.Nil)
	if err != nil && !isnil {
		// error
		isErr = true
		command := render(cmd.Args())
		msg = fmt.Sprintf("[Redis] %v | %s | %s", err, command, source)
		fields = logrus.Fields{"module": "redis", "command": command, "error": err, "source": source}
	} else {
		// result
		command := render(cmd.Args())
		rows, status := getDataFromCmder(cmd)
		first := "" // first field
		if isnil {
			first = "Nil" // Nil
		} else {
			first = status // OK | #
			if first == "" {
				first = strconv.Itoa(rows)
			}
		}
		msg = fmt.Sprintf("[Redis] %6s | %12s | %s | %s", first, duration.String(), command, source)
		fields = logrus.Fields{
			"module":   "redis",
			"command":  command,
			"rows":     rows,
			"status":   status,
			"duration": duration,
			"source":   source,
		}
	}

	return msg, fields, isErr
}

// render renders given command parameters to form the redis command expression.
func render(args []interface{}) string {
	sb := strings.Builder{}
	sb.WriteString(strings.ToUpper(args[0].(string)))
	for _, arg := range args[1:] {
		argStr := ""
		if s, ok := arg.(string); ok && strings.Contains(s, " ") {
			argStr = fmt.Sprintf("'%s'", s)
		} else {
			argStr = fmt.Sprintf("%v", arg)
		}
		sb.WriteRune(' ')
		sb.WriteString(argStr)
	}
	return sb.String()
}

// getDataFromCmder parses redis.Cmder and get rows and status.
func getDataFromCmder(cmder redis.Cmder) (rows int, status string) {
	switch cmd := cmder.(type) {
	// 1. value
	case *redis.StatusCmd:
		rows = 1
		status = cmd.Val()
	case *redis.BoolCmd:
		rows = 1
		status = "T"
		if !cmd.Val() {
			status = "F"
		}
	case *redis.IntCmd:
		rows = 1
		status = fmt.Sprintf("%di", cmd.Val())
	case *redis.FloatCmd:
		rows = 1
		status = fmt.Sprintf("%.1ff", cmd.Val())
	case *redis.StringCmd:
		rows = 1
		status = "str"

	// 2. slice
	case *redis.IntSliceCmd:
		rows = len(cmd.Val())
	case *redis.StringSliceCmd:
		rows = len(cmd.Val())
	case *redis.BoolSliceCmd:
		rows = len(cmd.Val())
	case *redis.SliceCmd:
		rows = len(cmd.Val())
	case *redis.ZSliceCmd:
		rows = len(cmd.Val())
	case *redis.XMessageSliceCmd:
		rows = len(cmd.Val())
	case *redis.XStreamSliceCmd:
		rows = len(cmd.Val())
	case *redis.ScanCmd:
		val, _ := cmd.Val()
		rows = len(val)
	case *redis.ClusterSlotsCmd:
		rows = len(cmd.Val())
	case *redis.GeoLocationCmd:
		rows = len(cmd.Val())
	case *redis.GeoPosCmd:
		rows = len(cmd.Val())
	case *redis.SlowLogCmd:
		rows = len(cmd.Val())
	case *redis.XInfoGroupsCmd:
		rows = len(cmd.Val())
	case *redis.XPendingCmd:
		c := cmd.Val().Consumers
		rows = len(c)
	case *redis.XPendingExtCmd:
		rows = len(cmd.Val())

	// 3. map
	case *redis.StringIntMapCmd:
		rows = len(cmd.Val())
	case *redis.StringStringMapCmd:
		rows = len(cmd.Val())
	case *redis.StringStructMapCmd:
		rows = len(cmd.Val())
	case *redis.CommandsInfoCmd:
		rows = len(cmd.Val())

	// 4. other
	case *redis.DurationCmd, *redis.TimeCmd, *redis.Cmd, *redis.XInfoStreamCmd, *redis.ZWithKeyCmd:
		rows = 1
	}
	return rows, status
}
