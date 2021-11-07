package xneo4j

import (
	"encoding/json"
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xstring"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/sirupsen/logrus"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// loggerOptions represents some options for logger, set by LoggerOption.
type loggerOptions struct {
	logErr       bool
	logCypher    bool
	skip         int
	counterField bool
}

// LoggerOption represents an option for logger, created by WithXXX functions.
type LoggerOption func(*loggerOptions)

// WithLogErr returns a LoggerOption with logErr switcher to do log for errors, defaults to true.
func WithLogErr(logErr bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logErr = logErr
	}
}

// WithLogCypher returns a LoggerOption with logCypher switcher to do log for cyphers, defaults to true.
func WithLogCypher(logCypher bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logCypher = logCypher
	}
}

// WithSkip returns a LoggerOption with runtime skip to get runtime information, defaults to 1.
func WithSkip(skip int) LoggerOption {
	return func(o *loggerOptions) {
		o.skip = skip
	}
}

// WithCounterField returns a LoggerOption with counter fields switcher, defaults to false.
func WithCounterField(switcher bool) LoggerOption {
	return func(o *loggerOptions) {
		o.counterField = switcher
	}
}

// _enable is a global switcher to control xneo4j logger behavior.
var _enable = true

// EnableLogger enables xneo4j logger to do any log.
func EnableLogger() {
	_enable = true
}

// DisableLogger disables xneo4j logger to do any log.
func DisableLogger() {
	_enable = false
}

// LogrusLogger represents a neo4j.Session, used to log neo4j cypher executing message to logrus.Logger.
type LogrusLogger struct {
	neo4j.Session
	logger  *logrus.Logger
	options *loggerOptions
}

var _ neo4j.Session = &LogrusLogger{}

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger and LoggerOption-s.
// Example:
// 	driver, err := neo4j.NewDriver(target, auth)
// 	session, err := driver.Session(neo4j.AccessModeRead)
// 	l := logrus.New()
// 	l.SetFormatter(&logrus.TextFormatter{})
// 	session = NewLogrusLogger(session, l) // with default skip 1
func NewLogrusLogger(session neo4j.Session, logger *logrus.Logger, options ...LoggerOption) *LogrusLogger {
	opt := &loggerOptions{
		logErr:       true,
		logCypher:    true,
		skip:         1,
		counterField: false,
	}
	for _, op := range options {
		if op != nil {
			op(opt)
		}
	}
	return &LogrusLogger{Session: session, logger: logger, options: opt}
}

// LoggerLogger represents a neo4j.Session, used to log neo4j cypher executing message to logrus.StdLogger.
type LoggerLogger struct {
	neo4j.Session
	logger  logrus.StdLogger
	options *loggerOptions
}

var _ neo4j.Session = &LoggerLogger{}

// NewLoggerLogger creates a new LoggerLogger using given log.Logger and LoggerOption-s.
// Example:
// 	driver, err := neo4j.NewDriver(target, auth)
// 	session, err := driver.Session(neo4j.AccessModeRead)
// 	l := log.New(os.Stderr, "", log.LstdFlags)
// 	session = NewLoggerLogger(session, l) // with default skip 1
func NewLoggerLogger(session neo4j.Session, logger logrus.StdLogger, options ...LoggerOption) *LoggerLogger {
	opt := &loggerOptions{
		logErr:       true,
		logCypher:    true,
		skip:         1,
		counterField: false,
	}
	for _, op := range options {
		if op != nil {
			op(opt)
		}
	}
	return &LoggerLogger{Session: session, logger: logger, options: opt}
}

// Run executes given cypher and params, and logs to logrus.Logger.
func (l *LogrusLogger) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error) {
	result, err := l.Session.Run(cypher, params, configurers...)
	if !_enable {
		return result, err
	}

	_, file, line, _ := runtime.Caller(l.options.skip)
	source := fmt.Sprintf("%s:%d", file, line)
	msg, fields, isErr := formatLoggerAndFields(result, err, source, l.options)
	if isErr && l.options.logErr {
		l.logger.WithFields(fields).Error(msg)
	} else if !isErr && l.options.logCypher {
		l.logger.WithFields(fields).Info(msg)
	}

	return result, err
}

// Run executes given cypher and params, and logs to logrus.StdLogger.
func (l *LoggerLogger) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error) {
	result, err := l.Session.Run(cypher, params, configurers...)
	if !_enable {
		return result, err
	}

	_, file, line, _ := runtime.Caller(l.options.skip)
	source := fmt.Sprintf("%s:%d", file, line)
	msg, _, isErr := formatLoggerAndFields(result, err, source, l.options)
	if (isErr && l.options.logErr) || (!isErr && l.options.logCypher) {
		l.logger.Print(msg)
	}

	return result, err
}

// formatLoggerAndFields formats given neo4j.Result, error, source and loggerOptions to logger string, logrus.Fields and isError flag.
// Logs like:
// 	[Neo4j] Connection error: dial tcp [::1]:7687: connectex: No connection could be made because the target machine actively refused it. | F:/Projects/ahlib-db/xneo4j/xneo4j_test.go:97
// 	[Neo4j] Server error: [Neo.ClientError.Statement.SyntaxError] Invalid input 'n' (line 1, column 26 (offset: 25)) | F:/Projects/ahlib-db/xneo4j/xneo4j_test.go:97
// 	[Neo4j]     -1 |        999ms | MATCH (n {uid: 8}) RETURN n LIMIT 1 | F:/Projects/ahlib-db/xneo4j/xneo4j_test.go:97
// 	       |------| |------------| |-----------------------------------| |---------------------------------------------|
// 	          6           12                        ...                                          ...
func formatLoggerAndFields(result neo4j.Result, err error, source string, options *loggerOptions) (string, logrus.Fields, bool) {
	var msg string
	var fields logrus.Fields
	var isErr bool

	if err != nil { // failed to connect (Connection error)
		// the target machine actively refused it
		// ...
		isErr = true
		fields = logrus.Fields{
			"module": "neo4j",
			"type":   "connection",
			"error":  err,
			"source": source,
		}
		msg = fmt.Sprintf("[Neo4j] %v | %s", err, source)
	} else if summary, err := result.Summary(); err != nil { // failed to execute (Server error)
		// Neo.ClientError.Security.Unauthorized
		// Neo.ClientError.Statement.SyntaxError
		// Neo.ClientError.Statement.TypeError
		// Neo.ClientError.Schema.ConstraintValidationFailed
		// ...
		isErr = true
		fields = logrus.Fields{
			"module": "neo4j",
			"type":   "server",
			"error":  err,
			"source": source,
		}
		msg = fmt.Sprintf("[Neo4j] %v | %s", err, source)
	} else {
		stat := summary.Statement()
		counters := summary.Counters()
		cypher := render(stat.Text(), stat.Params())
		du := summary.ResultAvailableAfter() + summary.ResultConsumedAfter()

		fields = logrus.Fields{
			"module":   "neo4j",
			"type":     "cypher",
			"cypher":   cypher,
			"rows":     -1, // unable to get rows, because this behavior will consume the iterator
			"duration": du,
			"source":   source,
		}
		// some statistics
		if options.counterField {
			fields["nodesCreated"] = counters.NodesCreated()
			fields["nodesDeleted"] = counters.NodesDeleted()
			fields["relationshipsCreated"] = counters.RelationshipsCreated()
			fields["relationshipsDeleted"] = counters.RelationshipsDeleted()
			fields["propertiesSet"] = counters.PropertiesSet()
			fields["labelsAdded"] = counters.LabelsAdded()
			fields["labelsRemoved"] = counters.LabelsRemoved()
			fields["indexesAdded"] = counters.IndexesAdded()
			fields["indexesRemoved"] = counters.IndexesRemoved()
			fields["constraintsAdded"] = counters.ConstraintsAdded()
			fields["constraintsRemoved"] = counters.ConstraintsRemoved()
		}
		msg = fmt.Sprintf("[Neo4j] %6d | %12s | %s | %s", -1, du, cypher, source)
	}

	return msg, fields, isErr
}

// render renders cypher string and parameters to complete cypher expression.
func render(cypher string, params map[string]interface{}) string {
	values := make(map[string]string, len(params))
	for k, v := range params {
		indirectValue := reflect.Indirect(reflect.ValueOf(v))
		if !indirectValue.IsValid() {
			values[k] = "NULL"
			continue
		}

		v = indirectValue.Interface()
		switch value := v.(type) {
		// Temporal values: https://neo4j.com/docs/cypher-manual/current/syntax/temporal/
		case neo4j.Date:
			values[k] = fmt.Sprintf(`date("%s")`, value.String())
		case neo4j.OffsetTime:
			values[k] = fmt.Sprintf(`time("%s")`, value.String())
		case time.Time:
			values[k] = fmt.Sprintf(`datetime("%s")`, value.Format(time.RFC3339Nano))
		case neo4j.LocalTime:
			values[k] = fmt.Sprintf(`localtime("%s")`, value.String())
		case neo4j.LocalDateTime:
			values[k] = fmt.Sprintf(`localdatetime("%s")`, value.String())
		case neo4j.Duration:
			values[k] = fmt.Sprintf(`duration("%s")`, value.String())

		// Spatial values: https://neo4j.com/docs/cypher-manual/3.5/syntax/spatial/
		case neo4j.Point:
			values[k] = value.String()

		// Slice and map values:
		case []interface{}:
			bs, err := json.Marshal(value)
			if err != nil {
				values[k] = "[?]"
			} else {
				values[k] = xstring.FastBtos(bs)
			}
		case map[string]interface{}:
			bs, err := json.Marshal(value)
			if err != nil {
				values[k] = "{?}"
			} else {
				values[k] = xstring.FastBtos(bs)
			}

		// other types
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
			values[k] = fmt.Sprintf("%v", value)
		default:
			values[k] = fmt.Sprintf("'%v'", value)
		}
	}

	result := cypher
	for k, v := range values {
		placeholder := fmt.Sprintf(`\$%s([^\w]|$)`, k) // `\$s([^\w]|$)` || `\$s(\W)`
		result = regexp.MustCompile(placeholder).ReplaceAllString(result, v+"$1")
	}
	return strings.TrimSpace(result)
}
