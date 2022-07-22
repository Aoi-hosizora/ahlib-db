package xredis

import (
	"context"
	"errors"
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xcolor"
	"github.com/Aoi-hosizora/ahlib/xstring"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

// loggerOptions is a type of LogrusLogger and StdLogger's option, each field can be set by LoggerOption function type.
type loggerOptions struct {
	logErr        bool
	logCmd        bool
	skip          int
	slowThreshold time.Duration
}

// LoggerOption represents an option type for LogrusLogger and StdLogger's option, can be created by WithXXX functions.
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

// WithSkip creates a LoggerOption to specify runtime skip for getting runtime information, defaults to 4.
func WithSkip(skip int) LoggerOption {
	return func(o *loggerOptions) {
		o.skip = skip
	}
}

// WithSlowThreshold creates a LoggerOption to specify a slow operation's duration used to highlight loggers, defaults to 0ms, means no highlight.
func WithSlowThreshold(threshold time.Duration) LoggerOption {
	return func(o *loggerOptions) {
		o.slowThreshold = threshold
	}
}

// _enable is a global flag to control behaviors of LogrusLogger and StdLogger, initials to true.
var _enable atomic.Value

func init() {
	_enable.Store(true)
}

// EnableLogger enables LogrusLogger and StdLogger to do any log.
func EnableLogger() {
	_enable.Store(true)
}

// DisableLogger disables LogrusLogger and StdLogger.
func DisableLogger() {
	_enable.Store(false)
}

// =====
// types
// =====

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
	logger *logrus.Logger
	option *loggerOptions
}

var _ redis.Hook = (*LogrusLogger)(nil)

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
func (l *StdLogger) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, &_startTimeKey, time.Now()), nil
}

// AfterProcess implements redis.Hook interface, it logs redis's message to logrus.Logger.
func (l *LogrusLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	endTime := time.Now()
	startTime, ok := ctx.Value(&_startTimeKey).(time.Time)
	enable := _enable.Load().(bool)
	if !enable || !ok {
		return nil // ignore
	}

	du := endTime.Sub(startTime)
	_, file, line, _ := runtime.Caller(l.option.skip) // defaults to 1
	source := fmt.Sprintf("%s:%d", file, line)

	p := extractLoggerParam(cmd, du, source, l.option)
	m := formatLoggerParam(p)
	f := fieldifyLoggerParam(p)
	if p.ErrorMsg != "" && l.option.logErr {
		l.logger.WithFields(f).Error(m)
	} else if p.ErrorMsg == "" && l.option.logCmd {
		l.logger.WithFields(f).Info(m)
	}
	return nil
}

// AfterProcess implements redis.Hook interface, it logs redis's message to logrus.StdLogger.
func (l *StdLogger) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	endTime := time.Now()
	startTime, ok := ctx.Value(&_startTimeKey).(time.Time)
	enable := _enable.Load().(bool)
	if !enable || !ok {
		return nil // ignore
	}

	du := endTime.Sub(startTime)
	_, file, line, _ := runtime.Caller(l.option.skip) // defaults to 1
	source := fmt.Sprintf("%s:%d", file, line)

	p := extractLoggerParam(cmd, du, source, l.option)
	m := formatLoggerParam(p)
	if (p.ErrorMsg != "" && l.option.logErr) || (p.ErrorMsg == "" && l.option.logCmd) {
		l.logger.Print(m)
	}
	return nil
}

// BeforeProcessPipeline implements redis.Hook interface.
func (*LogrusLogger) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

// AfterProcessPipeline implements redis.Hook interface.
func (*LogrusLogger) AfterProcessPipeline(context.Context, []redis.Cmder) error {
	return nil
}

// BeforeProcessPipeline implements redis.Hook interface.
func (*StdLogger) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

// AfterProcessPipeline implements redis.Hook interface.
func (*StdLogger) AfterProcessPipeline(context.Context, []redis.Cmder) error {
	return nil
}

// =======
// extract
// =======

// LoggerParam stores some logger parameters and is used by LogrusLogger and StdLogger.
type LoggerParam struct {
	Command  string
	Rows     int64
	Status   string
	Duration time.Duration
	Slow     bool
	Source   string
	ErrorMsg string
}

var (
	// FormatLoggerFunc is a custom LoggerParam's format function for LogrusLogger and StdLogger.
	FormatLoggerFunc func(p *LoggerParam) string

	// FieldifyLoggerFunc is a custom LoggerParam's fieldify function for LogrusLogger.
	FieldifyLoggerFunc func(p *LoggerParam) logrus.Fields
)

// extractLoggerParam extracts and returns LoggerParam using given parameters.
func extractLoggerParam(cmd redis.Cmder, duration time.Duration, source string, options *loggerOptions) *LoggerParam {
	err := cmd.Err()
	isnil := errors.Is(err, redis.Nil)
	command := render(cmd.Args())
	if err != nil && !isnil {
		return &LoggerParam{Command: command, ErrorMsg: err.Error(), Source: source}
	}
	rows, status := int64(0), "Nil"
	if !isnil {
		rows, status = statusFromCmder(cmd)
	}
	slow := options.slowThreshold > 0 && duration >= options.slowThreshold
	return &LoggerParam{
		Command:  command,
		Rows:     rows,
		Status:   status,
		Duration: duration,
		Slow:     slow,
		Source:   source,
	}
}

// formatLoggerParam formats given LoggerParam to string for LogrusLogger and StdLogger.
//
// The default format logs like:
// 	[Redis] err: ERR invalid password | SET test_a test_aaa | F:/Projects/ahlib-db/xredis/redis_test.go:41
// 	[Redis]    Nil |   305.9909ms | GET test | F:/Projects/ahlib-db/xredis/redis_test.go:126
// 	[Redis]      3 |      995.5µs | KEYS tes* | F:/Projects/ahlib-db/xredis/redis_test.go:146
// 	[Redis]     2i |      997.8µs | DEL test_ test_c | F:/Projects/ahlib-db/xredis/helper.go:21
// 	[Redis]      F |      997.3µs | HEXISTS test xxx | F:/Projects/ahlib-db/xredis/redis_test.go:141
// 	[Redis]     OK |    25.9306ms | SET 'test num' 1 | F:/Projects/ahlib-db/xredis/redis_test.go:59
// 	       |------| |------------| |----------------| |--------------------------------------------|
// 	          6           12               ...                               ...
func formatLoggerParam(p *LoggerParam) string {
	if FormatLoggerFunc != nil {
		return FormatLoggerFunc(p)
	}
	if p.ErrorMsg != "" {
		return fmt.Sprintf("[Redis] err: %v | %s | %s", p.ErrorMsg, p.Command, p.Source)
	}
	first := p.Status // # | status (OK, Nil, T, F, Str, ...)
	if first == "" {
		first = fmt.Sprintf("#=%d", p.Rows)
	}
	du := fmt.Sprintf("%12s", p.Duration.String())
	if p.Slow {
		du = xcolor.Yellow.WithStyle(xcolor.Bold).Sprintf("%12s", p.Duration.String())
	}
	return fmt.Sprintf("[Redis] %6s | %12s | %s | %s", first, du, p.Command, p.Source)
}

// fieldifyLoggerParam fieldifies given LoggerParam to logrus.Fields for LogrusLogger.
//
// The default contains the following fields: module, command, rows, status, duration, source, error.
func fieldifyLoggerParam(p *LoggerParam) logrus.Fields {
	if FieldifyLoggerFunc != nil {
		return FieldifyLoggerFunc(p)
	}
	if p.ErrorMsg != "" {
		return logrus.Fields{
			"module":  "redis",
			"command": p.Command,
			"error":   p.ErrorMsg,
			"source":  p.Source,
		}
	}
	return logrus.Fields{
		"module":   "redis",
		"command":  p.Command,
		"rows":     p.Rows,
		"status":   p.Status,
		"duration": p.Duration,
		"source":   p.Source,
	}
}

// ========
// internal
// ========

// render renders given command parameters to form the redis command expression.
func render(args []interface{}) string {
	sb := strings.Builder{}
	sb.WriteString(strings.ToUpper(args[0].(string)))
	for _, arg := range args[1:] {
		cmd := ""
		if argStr, ok := arg.(string); ok && strings.Contains(argStr, " ") {
			cmd = fmt.Sprintf("'%s'", argStr)
		} else {
			cmd = fmt.Sprintf("%v", arg)
		}
		sb.WriteRune(' ')
		sb.WriteString(cmd)
	}
	return sb.String()
}

// statusFromCmder parses redis.Cmder and returns rows and status information.
func statusFromCmder(cmder redis.Cmder) (rows int64, status string) {
	r, s := 1, ""
	switch cmd := cmder.(type) {
	// 1. value
	case *redis.StatusCmd:
		s = cmd.Val()
	case *redis.BoolCmd:
		s = xstring.Bool(cmd.Val(), "T", "F")
	case *redis.IntCmd:
		s = fmt.Sprintf("%di", cmd.Val())
	case *redis.FloatCmd:
		s = fmt.Sprintf("%.1ff", cmd.Val())
	case *redis.StringCmd:
		s = "Str"

	// 2. slice
	case *redis.IntSliceCmd:
		r = len(cmd.Val())
	case *redis.StringSliceCmd:
		r = len(cmd.Val())
	case *redis.BoolSliceCmd:
		r = len(cmd.Val())
	case *redis.SliceCmd:
		r = len(cmd.Val())
	case *redis.ZSliceCmd:
		r = len(cmd.Val())
	case *redis.XMessageSliceCmd:
		r = len(cmd.Val())
	case *redis.XStreamSliceCmd:
		r = len(cmd.Val())
	case *redis.ScanCmd:
		val, _ := cmd.Val()
		r = len(val)
	case *redis.ClusterSlotsCmd:
		r = len(cmd.Val())
	case *redis.GeoLocationCmd:
		r = len(cmd.Val())
	case *redis.GeoPosCmd:
		r = len(cmd.Val())
	case *redis.SlowLogCmd:
		r = len(cmd.Val())
	case *redis.XInfoGroupsCmd:
		r = len(cmd.Val())
	case *redis.XPendingCmd:
		r = len(cmd.Val().Consumers)
	case *redis.XPendingExtCmd:
		r = len(cmd.Val())

	// 3. map
	case *redis.StringIntMapCmd:
		r = len(cmd.Val())
	case *redis.StringStringMapCmd:
		r = len(cmd.Val())
	case *redis.StringStructMapCmd:
		r = len(cmd.Val())
	case *redis.CommandsInfoCmd:
		r = len(cmd.Val())

	// 4. other
	case *redis.DurationCmd,
		*redis.TimeCmd,
		*redis.Cmd,
		*redis.XInfoStreamCmd,
		*redis.ZWithKeyCmd:
		r = 1
	}
	return int64(r), s
}
