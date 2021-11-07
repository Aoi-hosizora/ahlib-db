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
	"time"
)

// ILogger represents neo4j's internal logger interface.
type ILogger interface {
	Printf(ctx context.Context, format string, v ...interface{})
}

// loggerOptions represents some options for logger, set by LoggerOption.
type loggerOptions struct {
	logErr bool
	logCmd bool
}

// LoggerOption represents an option for logger, created by WithXXX functions.
type LoggerOption func(*loggerOptions)

// WithLogErr returns a LoggerOption with logErr switcher to do log for errors, defaults to true.
func WithLogErr(logErr bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logErr = logErr
	}
}

// WithLogCmd returns a LoggerOption with logCmd switcher to do log for commands, defaults to true.
func WithLogCmd(logCmd bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logCmd = logCmd
	}
}

// _enable is a global switcher to control xredis logger behavior.
var _enable = true

// EnableLogger enables xredis logger to do any log.
func EnableLogger() {
	_enable = true
}

// DisableLogger disables xredis logger to do any log.
func DisableLogger() {
	_enable = false
}

// SilenceLogger represents a redis's logger, used to hide go-redis's info logger.
type SilenceLogger struct{}

// NewSilenceLogger creates a new SilenceLogger.
// Example:
// 	client := redis.NewClient(options)
// 	redis.SetLogger(xredis.NewSilenceLogger())
func NewSilenceLogger() *SilenceLogger {
	return &SilenceLogger{}
}

// Printf does nothing.
func (l *SilenceLogger) Printf(context.Context, string, ...interface{}) {}

// &_startTimeKey is the key for process start time
var _startTimeKey int

// LogrusLogger represents a redis's logger (as redis.Hook), used to log redis command executing message to logrus.Logger.
type LogrusLogger struct {
	logger  *logrus.Logger
	options *loggerOptions
}

var _ redis.Hook = &LogrusLogger{}

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger and LoggerOption-s.
// Example:
// 	client := redis.NewClient(options)
// 	redis.SetLogger(NewSilenceLogger())
// 	l := logrus.New()
// 	l.SetFormatter(&logrus.TextFormatter{})
// 	client.AddHook(xredis.NewLogrusLogger(l))
func NewLogrusLogger(logger *logrus.Logger, options ...LoggerOption) *LogrusLogger {
	opt := &loggerOptions{
		logErr: true,
		logCmd: true,
	}
	for _, op := range options {
		if op != nil {
			op(opt)
		}
	}
	return &LogrusLogger{logger: logger, options: opt}
}

// LoggerLogger represents a redis's logger (as redis.Hook), used to log redis command executing message to logrus.StdLogger.
type LoggerLogger struct {
	logger  logrus.StdLogger
	options *loggerOptions
}

var _ redis.Hook = &LoggerLogger{}

// NewLoggerLogger creates a new LoggerLogger using given logrus.StdLogger and LoggerOption-s.
// Example:
// 	client := redis.NewClient(options)
// 	redis.SetLogger(NewSilenceLogger())
// 	l := log.New(os.Stderr, "", log.LstdFlags)
// 	client.AddHook(xredis.NewLoggerLogger(l))
func NewLoggerLogger(logger logrus.StdLogger, options ...LoggerOption) *LoggerLogger {
	opt := &loggerOptions{
		logErr: true,
		logCmd: true,
	}
	for _, op := range options {
		if op != nil {
			op(opt)
		}
	}
	return &LoggerLogger{logger: logger, options: opt}
}

func (l *LogrusLogger) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (l *LogrusLogger) AfterProcessPipeline(context.Context, []redis.Cmder) error {
	return nil
}

// BeforeProcess saves start time to context, used in AfterProcess.
func (l *LogrusLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, &_startTimeKey, time.Now()), nil
}

// AfterProcess logs to logrus.Logger.
func (l *LogrusLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	endTime := time.Now()
	startTime, ok := ctx.Value(&_startTimeKey).(time.Time)
	if !_enable || !ok {
		return nil // ignore
	}

	_, file, line, _ := runtime.Caller(4)
	source := fmt.Sprintf("%s:%d", file, line)
	msg, fields, isErr := formatLoggerAndFields(cmd, endTime.Sub(startTime), source)
	if isErr && l.options.logErr {
		l.logger.WithFields(fields).Error(msg)
	} else if !isErr && l.options.logCmd {
		l.logger.WithFields(fields).Info(msg)
	}

	return nil
}

func (l *LoggerLogger) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (l *LoggerLogger) AfterProcessPipeline(context.Context, []redis.Cmder) error {
	return nil
}

// BeforeProcess saves start time to context, used in AfterProcess.
func (l *LoggerLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, &_startTimeKey, time.Now()), nil
}

// AfterProcess logs to logrus.StdLogger.
func (l *LoggerLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	endTime := time.Now()
	startTime, ok := ctx.Value(&_startTimeKey).(time.Time)
	if !_enable || !ok {
		return nil // ignore
	}

	_, file, line, _ := runtime.Caller(4)
	source := fmt.Sprintf("%s:%d", file, line)
	msg, _, isErr := formatLoggerAndFields(cmd, endTime.Sub(startTime), source)
	if (isErr && l.options.logErr) || (!isErr && l.options.logCmd) {
		l.logger.Print(msg)
	}

	return nil
}

// formatLoggerAndFields formats redis.Cmder and time.Duration to logger string, logrus.Fields and isError flag.
// Logs like:
// 	[Redis] ERR invalid password | SET test_a test_aaa | F:/Projects/ahlib-db/xredis/redis_test.go:41
// 	[Redis]    Nil |   305.9909ms | GET test | F:/Projects/ahlib-db/xredis/redis_test.go:126
// 	[Redis]      3 |      995.5µs | KEYS tes* | F:/Projects/ahlib-db/xredis/redis_test.go:146
// 	[Redis]     2i |      997.8µs | DEL test_ test_c | F:/Projects/ahlib-db/xredis/helper.go:21
// 	[Redis]      F |      997.3µs | HEXISTS test xxx | F:/Projects/ahlib-db/xredis/redis_test.go:141
// 	[Redis]     OK |    25.9306ms | SET 'test num' 1 | F:/Projects/ahlib-db/xredis/redis_test.go:59
// 	       |------| |------------| |----------------| |--------------------------------------------|
// 	          6           12               ...                               ...
func formatLoggerAndFields(cmd redis.Cmder, duration time.Duration, source string) (string, logrus.Fields, bool) {
	var msg string
	var fields logrus.Fields
	var isErr bool

	err := cmd.Err()
	isnil := errors.Is(err, redis.Nil)
	if err != nil && !isnil {
		// error
		isErr = true
		command := render(cmd.Args())
		fields = logrus.Fields{
			"module":  "redis",
			"command": command,
			"error":   err,
			"source":  source,
		}
		msg = fmt.Sprintf("[Redis] %v | %s | %s", err, command, source)
	} else {
		// result
		command := render(cmd.Args())
		rows, status := parseCmd(cmd)
		first := "" // first field
		if isnil {
			first = "Nil" // Nil
		} else {
			first = status // OK | #
			if first == "" {
				first = strconv.Itoa(rows)
			}
		}

		fields = logrus.Fields{
			"module":   "redis",
			"command":  command,
			"rows":     rows,
			"status":   status,
			"duration": duration,
			"source":   source,
		}
		msg = fmt.Sprintf("[Redis] %6s | %12s | %s | %s", first, duration.String(), command, source)
	}

	return msg, fields, isErr
}

// render renders command parameters to complete redis command expression.
func render(args []interface{}) string {
	sp := strings.Builder{}
	sp.WriteString(strings.ToUpper(args[0].(string)))
	for _, arg := range args[1:] {
		argStr := ""
		if s, ok := arg.(string); ok && strings.Contains(s, " ") {
			argStr = fmt.Sprintf("'%s'", s)
		} else {
			argStr = fmt.Sprintf("%v", arg)
		}
		sp.WriteRune(' ')
		sp.WriteString(argStr)
	}
	return sp.String()
}

// parseCmd parses redis.Cmder to each cmd types and get rows and status if the cmd is redis.StatusCmd.
func parseCmd(cmd redis.Cmder) (rows int, status string) {
	switch cmd := cmd.(type) {
	// slice
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
		k, _ := cmd.Val()
		rows = len(k)
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

	// map
	case *redis.StringIntMapCmd:
		rows = len(cmd.Val())
	case *redis.StringStringMapCmd:
		rows = len(cmd.Val())
	case *redis.StringStructMapCmd:
		rows = len(cmd.Val())
	case *redis.CommandsInfoCmd:
		rows = len(cmd.Val())

	// value
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

	// other
	case *redis.StringCmd, *redis.DurationCmd, *redis.TimeCmd, *redis.Cmd, *redis.XInfoStreamCmd, *redis.ZWithKeyCmd:
		rows = 1
	}
	return rows, status
}
