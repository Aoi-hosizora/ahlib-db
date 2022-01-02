package xgorm

import (
	"database/sql/driver"
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xstring"
	"github.com/sirupsen/logrus"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

// loggerOptions is a type of LogrusLogger and StdLogger's option, each field can be set by LoggerOption function type.
type loggerOptions struct {
	logInfo  bool
	logSQL   bool
	logOther bool
}

// LoggerOption represents an option type for LogrusLogger and StdLogger's option, can be created by WithXXX functions.
type LoggerOption func(*loggerOptions)

// WithLogInfo creates a LoggerOption to decide whether to do log for [INFO] or not, defaults to true.
func WithLogInfo(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logInfo = log
	}
}

// WithLogSQL creates a LoggerOption to decide whether to do log for [SQL] or not, defaults to true.
func WithLogSQL(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logSQL = log
	}
}

// WithLogOther creates a LoggerOption to decide whether to do log for other type (such as [LOG]) or not, defaults to true.
func WithLogOther(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logOther = log
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

// ILogger abstracts gorm's internal logger to an interface, equals to gorm.logger interface.
type ILogger interface {
	Print(v ...interface{})
}

// SilenceLogger represents a gorm's logger, used to hide all logs (such as "SQL" and "INFO"). Note that `gorm.DB.LogMode(false)` will only hide "SQL" message.
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
	_enableMu.RLock()
	enable := _enable
	_enableMu.RUnlock()
	if !enable || len(v) <= 1 {
		return // ignore
	}

	// [INFO] or [SQL] or [LOG] or ...
	m, f := extractLoggerData(v, l.option)
	if m != "" {
		l.logger.WithFields(f).Info(m)
	}
}

// Print implements gorm.logger interface, it logs gorm's message to logrus.StdLogger.
func (s *StdLogger) Print(v ...interface{}) {
	_enableMu.RLock()
	enable := _enable
	_enableMu.RUnlock()
	if !enable || len(v) <= 1 {
		return // ignore
	}

	// [INFO] or [SQL] or [LOG] or ...
	m, _ := extractLoggerData(v, s.option)
	if m != "" {
		s.logger.Print(m)
	}
}

// ========
// internal
// ========

// extractLoggerData extracts and formats given values to logger message and logrus.Fields, see gorm.LogFormatter for more details.
//
// Logs like:
// 	[Gorm] [info] registering callback `new_deleted_at_before_query_callback` from F:/Projects/ahlib-db/xgorm/hook.go:36
// 	[Gorm] [log] Error 1062: Duplicate entry '1' for key 'PRIMARY'
// 	[Gorm]       1 |     1.9957ms | SELECT * FROM `tbl_test`   ORDER BY `tbl_test`.`id` ASC LIMIT 1 | F:/Projects/ahlib-db/xgorm/xgorm_test.go:48
// 	      |-------| |------------| |---------------------------------------------------------------| |-------------------------------------------|
// 	          7           12                                      ...                                                       ...
func extractLoggerData(v []interface{}, option *loggerOptions) (string, logrus.Fields) {
	var msg string
	var fields logrus.Fields

	if len(v) == 2 {
		// [INFO]
		if !option.logInfo {
			return "", nil
		}
		msg = fmt.Sprintf("[Gorm] %v", v[1])
		fields = logrus.Fields{"module": "gorm", "type": v[0], "info": v[1]}
	} else if v[0] != "sql" {
		// [LOG], other ...
		if !option.logOther {
			return "", nil
		}
		s := fmt.Sprint(v[2:]...)
		msg = fmt.Sprintf("[Gorm] [%v] %v", v[0], s)
		fields = logrus.Fields{"module": "gorm", "type": v[0], "message": s}
	} else {
		// [SQL]
		if !option.logSQL {
			return "", nil
		}
		source := v[1]
		duration := v[2].(time.Duration)
		sql := render(v[3].(string), v[4].([]interface{}))
		rows := v[5].(int64)
		msg = fmt.Sprintf("[Gorm] %7d | %12s | %s | %s", rows, duration, sql, source)
		fields = logrus.Fields{
			"module":   "gorm",
			"type":     "sql",
			"sql":      sql,
			"rows":     rows,
			"duration": duration,
			"source":   source,
		}
	}

	return msg, fields
}

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
