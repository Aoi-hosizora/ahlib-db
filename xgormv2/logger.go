package xgormv2

import (
	"context"
	"errors"
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xcolor"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"sync/atomic"
	"time"
)

// loggerOptions is a type of LogrusLogger and StdLogger's option, each field can be set by LoggerOption function type.
type loggerOptions struct {
	logInfo       bool
	logSQL        bool
	logOther      bool
	slowThreshold time.Duration
}

// LoggerOption represents an option type for LogrusLogger and StdLogger's option, can be created by WithXXX functions.
type LoggerOption func(*loggerOptions)

// WithLogInfo creates a LoggerOption to decide whether to do log for [info] or not, defaults to true.
func WithLogInfo(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logInfo = log
	}
}

// WithLogSQL creates a LoggerOption to decide whether to do log for [sql] or not, defaults to true.
func WithLogSQL(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logSQL = log
	}
}

// WithLogOther creates a LoggerOption to decide whether to do log for other type (such as [warn] and [erro]) or not, defaults to true.
func WithLogOther(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logOther = log
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

// SilenceLogger represents a gorm's logger, used to hide all logs, including [info], [warn], [erro], [sql].
type SilenceLogger struct{}

var _ logger.Interface = (*SilenceLogger)(nil)

// NewSilenceLogger creates a new SilenceLogger.
//
// Example:
// 	gorm.Open(..., &gorm.Config{
// 		Logger: NewSilenceLogger(),
// 	})
func NewSilenceLogger() *SilenceLogger {
	return &SilenceLogger{}
}

func (s *SilenceLogger) LogMode(logger.LogLevel) logger.Interface                      { return s }
func (*SilenceLogger) Info(context.Context, string, ...interface{})                    {}
func (*SilenceLogger) Warn(context.Context, string, ...interface{})                    {}
func (*SilenceLogger) Error(context.Context, string, ...interface{})                   {}
func (*SilenceLogger) Trace(context.Context, time.Time, func() (string, int64), error) {}

// LogrusLogger represents a gorm's logger, used to log gorm's executing message to logrus.Logger.
type LogrusLogger struct {
	logger *logrus.Logger
	option *loggerOptions
}

var _ logger.Interface = (*LogrusLogger)(nil)

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger and LoggerOption-s.
//
// Example:
// 	l := logrus.New()
// 	l.SetFormatter(&logrus.TextFormatter{})
// 	gorm.Open(..., &gorm.Config{
// 		Logger: NewLogrusLogger(l),
// 	})
func NewLogrusLogger(logger *logrus.Logger, options ...LoggerOption) *LogrusLogger {
	opt := &loggerOptions{logInfo: true, logSQL: true, logOther: true}
	for _, o := range options {
		if o != nil {
			o(opt)
		}
	}
	return &LogrusLogger{logger: logger, option: opt}
}

// StdLogger represents a gorm's logger, used to log gorm's executing message to logrus.StdLogger.
type StdLogger struct {
	logger logrus.StdLogger
	option *loggerOptions
}

var _ logger.Interface = (*StdLogger)(nil)

// NewStdLogger creates a new StdLogger using given logrus.StdLogger and LoggerOption-s.
//
// Example:
// 	l := log.Default()
// 	gorm.Open(..., &gorm.Config{
// 		Logger: NewStdLogger(l),
// 	})
func NewStdLogger(logger logrus.StdLogger, options ...LoggerOption) *StdLogger {
	opt := &loggerOptions{logInfo: true, logSQL: true, logOther: true}
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

// doPrint implements gorm.logger interface, it logs gorm's message to logrus.Logger.
func (l *LogrusLogger) doPrint(level logger.LogLevel, msg string, v []interface{}) {
	enable := _enable.Load().(bool)
	if !enable {
		return // ignore
	}

	p := extractLoggerParam(level, msg, v, l.option)
	switch {
	case p == nil,
		p.Type == "info" && !l.option.logInfo,
		p.Type == "sql" && !l.option.logSQL,
		!l.option.logOther:
		return
	}
	m := formatLoggerParam(p)
	f := fieldifyLoggerParam(p)
	l.logger.WithFields(f).Info(m)
}

// doPrint implements gorm.logger interface, it logs gorm's message to logrus.StdLogger.
func (l *StdLogger) doPrint(level logger.LogLevel, msg string, v []interface{}) {
	enable := _enable.Load().(bool)
	if !enable {
		return // ignore
	}

	p := extractLoggerParam(level, msg, v, l.option)
	switch {
	case p == nil,
		p.Type == "info" && !l.option.logInfo,
		p.Type == "sql" && !l.option.logSQL,
		!l.option.logOther:
		return
	}
	m := formatLoggerParam(p)
	l.logger.Print(m)
}

// LogMode implements logger.Interface for LogrusLogger.
func (l *LogrusLogger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

// Info implements logger.Interface for LogrusLogger.
func (l *LogrusLogger) Info(_ context.Context, msg string, data ...interface{}) {
	l.doPrint(logger.Info, msg, data)
}

// Warn implements logger.Interface for LogrusLogger.
func (l *LogrusLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	l.doPrint(logger.Warn, msg, data)
}

// Error implements logger.Interface for LogrusLogger.
func (l *LogrusLogger) Error(_ context.Context, msg string, data ...interface{}) {
	l.doPrint(logger.Error, msg, data)
}

// Trace implements logger.Interface for LogrusLogger.
func (l *LogrusLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, row := fc()
	l.doPrint(logger.Info+1, "", []interface{}{utils.FileWithLineNum(), time.Now().Sub(begin), sql, row, err})
}

// LogMode implements logger.Interface for StdLogger.
func (l *StdLogger) LogMode(logger.LogLevel) logger.Interface {
	return l
}

// Info implements logger.Interface for StdLogger.
func (l *StdLogger) Info(_ context.Context, msg string, data ...interface{}) {
	l.doPrint(logger.Info, msg, data)
}

// Warn implements logger.Interface for StdLogger.
func (l *StdLogger) Warn(_ context.Context, msg string, data ...interface{}) {
	l.doPrint(logger.Warn, msg, data)
}

// Error implements logger.Interface for StdLogger.
func (l *StdLogger) Error(_ context.Context, msg string, data ...interface{}) {
	l.doPrint(logger.Error, msg, data)
}

// Trace implements logger.Interface for StdLogger.
func (l *StdLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, row := fc()
	l.doPrint(logger.Info+1, "", []interface{}{utils.FileWithLineNum(), time.Now().Sub(begin), sql, row, err})
}

// =======
// extract
// =======

// LoggerParam stores some logger parameters and is used by LogrusLogger and StdLogger.
type LoggerParam struct {
	Type     string
	Message  string
	SQL      string
	Rows     int64
	Duration time.Duration
	Slow     bool
	Source   string
}

var (
	// FormatLoggerFunc is a custom LoggerParam's format function for LogrusLogger and StdLogger.
	FormatLoggerFunc func(p *LoggerParam) string

	// FieldifyLoggerFunc is a custom LoggerParam's fieldify function for LogrusLogger.
	FieldifyLoggerFunc func(p *LoggerParam) logrus.Fields
)

// extractLoggerParam extracts and returns LoggerParam using given parameters.
func extractLoggerParam(level logger.LogLevel, msg string, v []interface{}, options *loggerOptions) *LoggerParam {
	switch {
	case level == logger.Info:
		return &LoggerParam{Type: "info", Message: fmt.Sprint("[info] ", fmt.Sprintf(msg, v...))}
	case level == logger.Warn:
		return &LoggerParam{Type: "warn", Message: fmt.Sprint("[warn] ", fmt.Sprintf(msg, v...))}
	case level == logger.Error:
		return &LoggerParam{Type: "erro", Message: fmt.Sprint("[erro] ", fmt.Sprintf(msg, v...))}
	case len(v) == 0:
		return nil // unreachable
	default: // [SQL]
		source := v[0].(string)
		duration := v[1].(time.Duration)
		slow := options.slowThreshold > 0 && duration >= options.slowThreshold
		sql := v[2].(string)
		rows := v[3].(int64)
		err, ok := v[4].(error)
		errMsg := ""
		if ok && !errors.Is(err, gorm.ErrRecordNotFound) {
			errMsg = err.Error()
		}
		return &LoggerParam{
			Type:     "sql",
			SQL:      sql,
			Rows:     rows,
			Duration: duration,
			Slow:     slow,
			Message:  errMsg,
			Source:   source,
		}
	}
}

// formatLoggerParam formats given LoggerParam to string for LogrusLogger and StdLogger.
//
// The default format logs like:
// 	[Gorm] [info] registering callback `new_deleted_at_before_query_callback` from F:/Projects/ahlib-db/xgorm/hook.go:36
// 	[Gorm] [log] Error 1062: Duplicate entry '1' for key 'PRIMARY'
// 	[Gorm]       1 |     1.9957ms | SELECT * FROM `tbl_test`   ORDER BY `tbl_test`.`id` ASC LIMIT 1 | F:/Projects/ahlib-db/xgorm/xgorm_test.go:48
// 	      |-------| |------------| |---------------------------------------------------------------| |-------------------------------------------|
// 	          7           12                                      ...                                                       ...
func formatLoggerParam(p *LoggerParam) string {
	if FormatLoggerFunc != nil {
		return FormatLoggerFunc(p)
	}
	if p.Type != "sql" {
		return fmt.Sprintf("[Gorm] %s", p.Message)
	}
	du := fmt.Sprintf("%12s", p.Duration.String())
	if p.Slow {
		du = xcolor.Yellow.WithStyle(xcolor.Bold).Sprintf("%12s", p.Duration.String())
	}
	foot := p.Source
	if p.Message != "" {
		foot = fmt.Sprintf("%s | %s", p.Message, p.Source)
	}
	return fmt.Sprintf("[Gorm] %7d | %s | %s | %s", p.Rows, du, p.SQL, foot)
}

// fieldifyLoggerParam fieldifies given LoggerParam to logrus.Fields for LogrusLogger.
//
// The default contains the following fields: module, type, message, sql, rows, duration, source.
func fieldifyLoggerParam(p *LoggerParam) logrus.Fields {
	if FieldifyLoggerFunc != nil {
		return FieldifyLoggerFunc(p)
	}
	if p.Type != "sql" {
		return logrus.Fields{
			"module":  "gorm",
			"type":    p.Type, // info / warn / erro
			"message": p.Message,
		}
	}
	return logrus.Fields{
		"module":   "gorm",
		"type":     p.Type, // sql
		"sql":      p.SQL,
		"rows":     p.Rows,
		"duration": p.Duration,
		"source":   p.Source,
		"message":  p.Message,
	}
}
