package xgorm

import (
	"database/sql/driver"
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xcolor"
	"github.com/Aoi-hosizora/ahlib/xstring"
	"github.com/sirupsen/logrus"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
	"unicode"
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

// WithLogOther creates a LoggerOption to decide whether to do log for other type (such as [log]) or not, defaults to true.
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

// ILogger abstracts gorm's internal logger to an interface, equals to gorm.logger interface.
type ILogger interface {
	Print(v ...interface{})
}

// SilenceLogger represents a gorm's logger, used to hide all logs, including [info], [sql] and so on. Note that `gorm.DB.LogMode(false)` will only hide [sql] message.
type SilenceLogger struct{}

// NewSilenceLogger creates a new SilenceLogger.
//
// Example:
// 	db, err := gorm.Open("mysql", dsn)
// 	db.LogMode(false) // both true and false are ok
// 	db.SetLogger(xgorm.NewSilenceLogger())
func NewSilenceLogger() *SilenceLogger {
	return &SilenceLogger{}
}

// Print implements gorm.logger interface, it does nothing for logging.
func (s *SilenceLogger) Print(...interface{}) {}

// LogrusLogger represents a gorm's logger, used to log gorm's executing message to logrus.Logger.
type LogrusLogger struct {
	logger *logrus.Logger
	option *loggerOptions
}

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger and LoggerOption-s.
//
// Example:
// 	db, err := gorm.Open("mysql", dsn)
// 	db.LogMode(true) // must be true
// 	l := logrus.New()
// 	l.SetFormatter(&logrus.TextFormatter{})
// 	db.SetLogger(xgorm.NewLogrusLogger(l))
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

// NewStdLogger creates a new StdLogger using given logrus.StdLogger and LoggerOption-s.
//
// Example:
// 	db, err := gorm.Open("mysql", dsn)
// 	db.LogMode(true) // must be true
// 	l := log.Default()
// 	db.SetLogger(xgorm.NewStdLogger(l))
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

// Print implements gorm.logger interface, it logs gorm's message to logrus.Logger.
func (l *LogrusLogger) Print(v ...interface{}) {
	enable := _enable.Load().(bool)
	if !enable || len(v) <= 1 {
		return // ignore
	}

	p := extractLoggerParam(v, l.option)
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

// Print implements gorm.logger interface, it logs gorm's message to logrus.StdLogger.
func (l *StdLogger) Print(v ...interface{}) {
	enable := _enable.Load().(bool)
	if !enable || len(v) <= 1 {
		return // ignore
	}

	p := extractLoggerParam(v, l.option)
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
func extractLoggerParam(v []interface{}, options *loggerOptions) *LoggerParam {
	switch {
	case len(v) <= 1:
		return nil // unreachable
	case len(v) == 2: // [INFO], ...
		return &LoggerParam{Type: v[0].(string), Message: fmt.Sprintf("%v", v[1])}
	case v[0] != "sql": // [LOG], ...
		return &LoggerParam{Type: v[0].(string), Message: fmt.Sprintf("[%s] %v", v[0], fmt.Sprint(v[2:]...))}
	default: // [SQL]
		source := v[1].(string)
		duration := v[2].(time.Duration)
		slow := options.slowThreshold > 0 && duration >= options.slowThreshold
		sql := render(v[3].(string), v[4].([]interface{}))
		rows := v[5].(int64)
		return &LoggerParam{
			Type:     v[0].(string),
			SQL:      sql,
			Rows:     rows,
			Duration: duration,
			Slow:     slow,
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
	return fmt.Sprintf("[Gorm] %7d | %s | %s | %s", p.Rows, du, p.SQL, p.Source)
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
			"type":    p.Type, // info / log / ...
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
	}
}

// ========
// internal
// ========

// isPrintable is a string utility function used in render.
func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

// some regexps used in render.
var (
	_placeholderRegexp        = regexp.MustCompile(`\?`)
	_numericPlaceholderRegexp = regexp.MustCompile(`\$\d+`)
)

// render renders given sql string and parameters to form the sql expression.
func render(sql string, params []interface{}) string {
	values := make([]string, 0, len(params))
	for _, v := range params {
		indirectValue := reflect.Indirect(reflect.ValueOf(v))
		if !indirectValue.IsValid() {
			values = append(values, "NULL")
			continue
		}

		v = indirectValue.Interface()
		switch value := v.(type) {
		case time.Time:
			if value.IsZero() {
				values = append(values, fmt.Sprintf("'%v'", "0000-00-00 00:00:00"))
			} else {
				values = append(values, fmt.Sprintf("'%v'", value.Format("2006-01-02 15:04:05")))
			}
		case []byte:
			if str := xstring.FastBtos(value); isPrintable(str) {
				values = append(values, fmt.Sprintf("'%v'", str))
			} else {
				values = append(values, "'<binary>'")
			}
		case driver.Valuer:
			if val, err := value.Value(); err == nil && val != nil {
				values = append(values, fmt.Sprintf("'%v'", val))
			} else {
				values = append(values, "NULL")
			}
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
			values = append(values, fmt.Sprintf("%v", value))
		default:
			values = append(values, fmt.Sprintf("'%v'", value))
		}
	}

	result := ""
	if _numericPlaceholderRegexp.MatchString(sql) { // \$\d+
		result = sql + " "
		for idx, value := range values {
			placeholder := fmt.Sprintf(`\$%d(\D)`, idx+1) // \$1(\D) X
			result = regexp.MustCompile(placeholder).ReplaceAllString(result, value+"$1")
		}
	} else {
		for idx, val := range _placeholderRegexp.Split(sql, -1) { // \?
			result += val
			if idx < len(values) {
				result += values[idx]
			}
		}
	}
	return strings.TrimSpace(result)
}
