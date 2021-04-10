package xgorm

import (
	"database/sql/driver"
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"
)

// ILogger represents gorm's internal logger interface.
type ILogger interface {
	Print(v ...interface{})
}

// loggerOptions represents some options for logger, set by LoggerOption.
type loggerOptions struct {
	logInfo  bool
	logOther bool
}

// LoggerOption represents an option for logger, created by WithXXX functions.
type LoggerOption func(*loggerOptions)

// WithLogInfo returns a LoggerOption with logInfo switcher to do log for [INFO], defaults to true.
func WithLogInfo(logInfo bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logInfo = logInfo
	}
}

// WithLogOther returns a LoggerOption with logOther switcher to do log for other type, such as [LOG], defaults to true.
func WithLogOther(logOther bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logOther = logOther
	}
}

// _enable is a global switcher to control xgorm logger behavior.
var _enable = true

// EnableLogger enables xgorm logger to do any log.
func EnableLogger() {
	_enable = true
}

// DisableLogger disables xgorm logger to do any log.
func DisableLogger() {
	_enable = false
}

// SilenceLogger represents a gorm's logger, used to hide "SQL" and "INFO" logs. Note that `gorm.DB.LogMode(false)` will only hide "SQL" message.
type SilenceLogger struct{}

// NewSilenceLogger creates a new SilenceLogger.
// Example:
// 	db, err := gorm.Open("mysql", dsl)
// 	db.LogMode(true) // both true and false are ok
// 	db.SetLogger(xgorm.NewSilenceLogger())
func NewSilenceLogger() *SilenceLogger {
	return &SilenceLogger{}
}

// LogrusLogger represents a gorm's logger, used to log "SQL" and "INFO" message to logrus.Logger.
type LogrusLogger struct {
	logger  *logrus.Logger
	options *loggerOptions
}

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger and LoggerOption-s.
// Example:
// 	db, err := gorm.Open("mysql", dsl)
// 	db.LogMode(true) // must be true
// 	l := logrus.New()
// 	l.SetFormatter(&logrus.TextFormatter{})
// 	db.SetLogger(xgorm.NewLogrusLogger(l))
func NewLogrusLogger(logger *logrus.Logger, options ...LoggerOption) *LogrusLogger {
	opt := &loggerOptions{
		logInfo:  true,
		logOther: true,
	}
	for _, op := range options {
		if op != nil {
			op(opt)
		}
	}
	return &LogrusLogger{logger: logger, options: opt}
}

// LoggerLogger represents a gorm's logger, used to log "SQL" and "INFO" message to logrus.StdLogger.
type LoggerLogger struct {
	logger  logrus.StdLogger
	options *loggerOptions
}

// NewLoggerLogger creates a new LoggerLogger using given logrus.StdLogger and LoggerOption-s.
// Example:
// 	db, err := gorm.Open("mysql", dsl)
// 	db.LogMode(true) // must be true
// 	l := log.New(os.Stderr, "", log.LstdFlags)
// 	db.SetLogger(xgorm.NewLoggerLogger(l))
func NewLoggerLogger(logger logrus.StdLogger, options ...LoggerOption) *LoggerLogger {
	opt := &loggerOptions{
		logInfo:  true,
		logOther: true,
	}
	for _, op := range options {
		if op != nil {
			op(opt)
		}
	}
	return &LoggerLogger{logger: logger, options: opt}
}

// Print does nothing for log.
func (g *SilenceLogger) Print(...interface{}) {}

// Print logs to logrus.Logger, see gorm.LogFormatter for details.
func (g *LogrusLogger) Print(v ...interface{}) {
	if !_enable || len(v) <= 1 {
		return
	}

	// info & sql & ...
	msg, fields := formatLoggerAndFields(v, g.options)
	if msg != "" && len(fields) != 0 {
		g.logger.WithFields(fields).Info(msg)
	}
}

// Print logs to logrus.StdLogger, see gorm.LogFormatter for details.
func (g *LoggerLogger) Print(v ...interface{}) {
	if !_enable || len(v) <= 1 {
		return
	}

	// info & sql & ...
	msg, _ := formatLoggerAndFields(v, g.options)
	if msg != "" {
		g.logger.Print(msg)
	}
}

// formatLoggerAndFields formats interface{}-s to logger string and logrus.Fields.
// Logs like:
// 	[Gorm] [info] registering callback `new_deleted_at_before_query_callback` from F:/Projects/ahlib-db/xgorm/hook.go:36
// 	[Gorm] [log] Error 1062: Duplicate entry '1' for key 'PRIMARY'
// 	[Gorm]       1 |     1.9957ms | SELECT * FROM `tbl_test`   ORDER BY `tbl_test`.`id` ASC LIMIT 1 | F:/Projects/ahlib-db/xgorm/xgorm_test.go:48
// 	      |-------| |------------| |---------------------------------------------------------------| |-------------------------------------------|
// 	          7           12                                      ...                                                       ...
func formatLoggerAndFields(v []interface{}, options *loggerOptions) (string, logrus.Fields) {
	var msg string
	var fields logrus.Fields

	if len(v) == 2 {
		// info
		if !options.logInfo {
			return "", nil
		}
		fields = logrus.Fields{
			"module": "gorm",
			"type":   v[0],
			"info":   v[1],
		}
		msg = fmt.Sprintf("[Gorm] %v", v[1])
	} else if v[0] != "sql" {
		// other
		if !options.logOther {
			return "", nil
		}
		s := fmt.Sprint(v[2:]...)
		fields = logrus.Fields{
			"module":  "gorm",
			"type":    v[0],
			"message": s,
		}
		msg = fmt.Sprintf("[Gorm] [%v] %v", v[0], s)
	} else {
		// sql
		source := v[1]
		duration := v[2].(time.Duration)
		sql := render(v[3].(string), v[4].([]interface{}))
		rows := v[5].(int64)

		fields = logrus.Fields{
			"module":   "gorm",
			"type":     "sql",
			"sql":      sql,
			"rows":     rows,
			"duration": duration,
			"source":   source,
		}
		msg = fmt.Sprintf("[Gorm] %7d | %12s | %s | %s", rows, duration, sql, source)
	}

	return msg, fields
}

// some regexps used in render.
var (
	_placeholderRegexp        = regexp.MustCompile(`\?`)
	_numericPlaceholderRegexp = regexp.MustCompile(`\$\d+`)
)

// isPrintable is a string util function used in render.
func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}

// render renders sql string and parameters to complete sql expression.
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
			if str := string(value); isPrintable(str) {
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
