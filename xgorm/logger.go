package xgorm

import (
	"database/sql/driver"
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// SilenceLogger hides "SQL" and "INFO" logs. Note that `gorm.DB.LogMode(false)` will only hide "SQL" message.
type SilenceLogger struct{}

// NewSilenceLogger creates a new SilenceLogger.
func NewSilenceLogger() *SilenceLogger {
	return &SilenceLogger{}
}

// LogrusLogger logs "SQL" and "INFO" message to logrus.Logger.
type LogrusLogger struct {
	logger *logrus.Logger
}

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger.
func NewLogrusLogger(logger *logrus.Logger) *LogrusLogger {
	return &LogrusLogger{logger: logger}
}

// LoggerLogger logs "SQL" and "INFO" message to logrus.StdLogger.
type LoggerLogger struct {
	logger logrus.StdLogger
}

// NewLoggerLogger creates a new LoggerLogger using given logrus.StdLogger.
func NewLoggerLogger(logger logrus.StdLogger) *LoggerLogger {
	return &LoggerLogger{logger: logger}
}

// Print does nothing for log.
func (g *SilenceLogger) Print(...interface{}) {}

// Print logs to logrus.Logger, see gorm.LogFormatter for details.
func (g *LogrusLogger) Print(v ...interface{}) {
	if len(v) <= 1 {
		return
	}

	// info & sql & ...
	msg, fields := formatLoggerAndFields(v)
	g.logger.WithFields(fields).Info(msg)
}

// Print logs to logrus.StdLogger, see gorm.LogFormatter for details.
func (g *LoggerLogger) Print(v ...interface{}) {
	if len(v) <= 1 {
		return
	}

	// info & sql & ...
	msg, _ := formatLoggerAndFields(v)
	g.logger.Print(msg)
}

// formatLoggerAndFields formats interface{}-s to logger string and logrus.Fields.
// Logs like:
// 	[Gorm] xxx
func formatLoggerAndFields(v []interface{}) (string, logrus.Fields) {
	var msg string
	var fields logrus.Fields

	if len(v) == 2 {
		// info
		fields = logrus.Fields{
			"module": "gorm",
			"type":   v[0],
			"info":   v[1],
		}
		msg = fmt.Sprintf("[Gorm] %v", v[1])
	} else if v[0] != "sql" {
		// other
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
		msg = fmt.Sprintf("[Gorm] #: %4d | %12s | %s | %s", rows, duration, sql, source)
	}

	return msg, fields
}

var (
	_sqlRegexp = regexp.MustCompile(`(\$\d+)|\?`) // used in render
)

// render renders sql string and parameters to complete sql expression.
func render(sql string, params []interface{}) string {
	values := make([]interface{}, 0)
	for _, value := range params {
		indirectValue := reflect.Indirect(reflect.ValueOf(value))
		if indirectValue.IsValid() { // valid
			value = indirectValue.Interface()
			if t, ok := value.(time.Time); ok { // time
				values = append(values, fmt.Sprintf("'%v'", t.Format(time.RFC3339)))
			} else if b, ok := value.([]byte); ok { // bytes
				values = append(values, fmt.Sprintf("'%v'", string(b)))
			} else if r, ok := value.(driver.Valuer); ok { // driver
				if value, err := r.Value(); err == nil && value != nil {
					values = append(values, fmt.Sprintf("'%v'", value))
				} else {
					values = append(values, "NULL")
				}
			} else { // other value
				values = append(values, fmt.Sprintf("'%v'", value))
			}
		} else { // invalid
			values = append(values, fmt.Sprintf("'%v'", value))
		}
	}

	result := fmt.Sprintf(_sqlRegexp.ReplaceAllString(sql, "%v"), values...)
	return strings.TrimSpace(result)
}
