package xredis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
	"time"
)

const (
	_startTimeKey = "XREDIS_PROCESS_START_TIME" // key for process start time
)

// LogrusLogger logs redis command executing message to logrus.Logger.
type LogrusLogger struct {
	logger *logrus.Logger
}

var _ redis.Hook = &LogrusLogger{}

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger.
func NewLogrusLogger(logger *logrus.Logger) *LogrusLogger {
	return &LogrusLogger{logger: logger}
}

// LoggerLogger logs redis command executing message to logrus.StdLogger.
type LoggerLogger struct {
	logger logrus.StdLogger
}

var _ redis.Hook = &LoggerLogger{}

// NewLoggerLogger creates a new LoggerLogger using given logrus.StdLogger.
func NewLoggerLogger(logger logrus.StdLogger) *LoggerLogger {
	return &LoggerLogger{logger: logger}
}

func (l *LogrusLogger) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (l *LogrusLogger) AfterProcessPipeline(context.Context, []redis.Cmder) error {
	return nil
}

// BeforeProcess saves start time to context, used in AfterProcess.
func (l *LogrusLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, _startTimeKey, time.Now()), nil
}

// AfterProcess logs to logrus.Logger.
func (l *LogrusLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	endTime := time.Now()
	startTime, ok := ctx.Value(_startTimeKey).(time.Time)
	if !ok {
		return nil // ignore
	}

	_, file, line, _ := runtime.Caller(4)
	source := fmt.Sprintf("%s:%d", file, line)
	msg, fields, isErr := formatLoggerAndFields(cmd, endTime.Sub(startTime), source)
	if isErr {
		l.logger.WithFields(fields).Error(msg)
	} else {
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
	return context.WithValue(ctx, _startTimeKey, time.Now()), nil
}

// AfterProcess logs to logrus.StdLogger.
func (l *LoggerLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	endTime := time.Now()
	startTime, ok := ctx.Value(_startTimeKey).(time.Time)
	if !ok {
		return nil // ignore
	}

	_, file, line, _ := runtime.Caller(4)
	source := fmt.Sprintf("%s:%d", file, line)
	msg, _, _ := formatLoggerAndFields(cmd, endTime.Sub(startTime), source)
	l.logger.Print(msg)

	return nil
}

// formatLoggerAndFields formats redis.Cmder and time.Duration to logger string, logrus.Fields and isError flag.
// Logs like:
// 	[Redis]      1 |     1.9968ms | get a | F:/Projects/ahlib-db/xredis/redis_test.go:82
// 	       |------| |------------| |-----| |--------------------------------------------|
// 	          6           12         ...                         ...
func formatLoggerAndFields(cmd redis.Cmder, duration time.Duration, source string) (string, logrus.Fields, bool) {
	var msg string
	var fields logrus.Fields
	var isErr bool

	if err := cmd.Err(); err != nil {
		// error
		isErr = true
		fields = logrus.Fields{
			"module": "redis",
			"error":  err,
			"source": source,
		}
		msg = fmt.Sprintf("[Redis] %v | %s", err, source)
	} else {
		// result
		sp := strings.Builder{}
		for _, arg := range cmd.Args() {
			if sp.Len() > 0 {
				sp.WriteRune(' ')
			}
			sp.WriteString(fmt.Sprintf("%v", arg))
		}
		command := sp.String()
		rows, status := parseCmd(cmd)

		fields = logrus.Fields{
			"module":   "redis",
			"command":  command,
			"rows":     rows,
			"status":   status,
			"duration": duration,
			"source":   source,
		}
		msg = fmt.Sprintf("[Redis] %6d | %12s | %s | %s", rows, duration.String(), command, source)
	}

	return msg, fields, isErr
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

	// others
	case *redis.StatusCmd:
		rows = 1
		status = cmd.Val()
	case *redis.BoolCmd, *redis.FloatCmd, *redis.IntCmd, *redis.StringCmd, *redis.DurationCmd, *redis.TimeCmd,
		*redis.Cmd, *redis.XInfoStreamCmd, *redis.ZWithKeyCmd:
		rows = 1
	default:
		rows = 1
	}
	return rows, status
}
