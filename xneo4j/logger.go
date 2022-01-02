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
	"sync"
	"time"
)

// loggerOptions is a type of LogrusLogger and StdLogger's option, each field can be set by LoggerOption function type.
type loggerOptions struct {
	logErr       bool
	logCypher    bool
	counterField bool
	skip         int
}

// LoggerOption represents an option type for LogrusLogger and StdLogger's option, can be created by WithXXX functions.
type LoggerOption func(*loggerOptions)

// WithLogErr creates a LoggerOption to decide whether to do log for errors or not, defaults to true.
func WithLogErr(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logErr = log
	}
}

// WithLogCypher creates a LoggerOption to decide whether to do log for cyphers or not, defaults to true.
func WithLogCypher(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logCypher = log
	}
}

// WithCounterField creates a LoggerOption to decide whether to do store counter fields to logrus.Fields or not, defaults to false.
func WithCounterField(flag bool) LoggerOption {
	return func(o *loggerOptions) {
		o.counterField = flag
	}
}

// WithSkip creates a LoggerOption to specific runtime skip for getting runtime information, defaults to 1.
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

// LogrusLogger represents a neo4j.Session as neo4j's logger, used to log neo4j's executing message to logrus.Logger.
type LogrusLogger struct {
	neo4j.Session
	logger *logrus.Logger
	option *loggerOptions
}

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger and LoggerOption-s.
//
// Example:
// 	driver, err := neo4j.NewDriver(...)
// 	sess, err := driver.Session(neo4j.AccessModeRead)
// 	l := logrus.New()
// 	l.SetFormatter(&logrus.TextFormatter{})
// 	sess = NewLogrusLogger(sess, l)
func NewLogrusLogger(session neo4j.Session, logger *logrus.Logger, options ...LoggerOption) *LogrusLogger {
	opt := &loggerOptions{logErr: true, logCypher: true, counterField: false, skip: 1}
	for _, op := range options {
		if op != nil {
			op(opt)
		}
	}
	return &LogrusLogger{Session: session, logger: logger, option: opt}
}

// StdLogger represents a neo4j.Session as neo4j's logger, used to log neo4j's executing message to logrus.StdLogger.
type StdLogger struct {
	neo4j.Session
	logger logrus.StdLogger
	option *loggerOptions
}

// NewStdLogger creates a new StdLogger using given log.Logger and LoggerOption-s.
//
// Example:
// 	driver, err := neo4j.NewDriver(...)
// 	sess, err := driver.Session(neo4j.AccessModeRead)
// 	l := log.Default()
// 	sess = NewStdLogger(sess, l)
func NewStdLogger(session neo4j.Session, logger logrus.StdLogger, options ...LoggerOption) *StdLogger {
	opt := &loggerOptions{logErr: true, logCypher: true, counterField: false, skip: 1}
	for _, op := range options {
		if op != nil {
			op(opt)
		}
	}
	return &StdLogger{Session: session, logger: logger, option: opt}
}

// =======
// methods
// =======

// Run implements neo4j.Session interface, it executes given cypher and params, and logs neo4j's message to logrus.Logger.
func (l *LogrusLogger) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error) {
	result, err := l.Session.Run(cypher, params, configurers...)
	_enableMu.RLock()
	enable := _enable
	_enableMu.RUnlock()
	if !enable {
		return result, err // ignore
	}

	_, file, line, _ := runtime.Caller(l.option.skip) // defaults to 4
	source := fmt.Sprintf("%s:%d", file, line)
	m, f, isErr := extractLoggerData(result, err, source, l.option)
	if isErr && l.option.logErr {
		l.logger.WithFields(f).Error(m)
	} else if !isErr && l.option.logCypher {
		l.logger.WithFields(f).Info(m)
	}

	return result, err
}

// Run implements neo4j.Session interface, it executes given cypher and params, and logs neo4j's message to logrus.StdLogger.
func (s *StdLogger) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error) {
	result, err := s.Session.Run(cypher, params, configurers...)
	_enableMu.RLock()
	enable := _enable
	_enableMu.RUnlock()
	if !enable {
		return result, err // ignore
	}

	_, file, line, _ := runtime.Caller(s.option.skip) // defaults to 4
	source := fmt.Sprintf("%s:%d", file, line)
	m, _, isErr := extractLoggerData(result, err, source, s.option)
	if (isErr && s.option.logErr) || (!isErr && s.option.logCypher) {
		s.logger.Print(m)
	}

	return result, err
}

// ========
// internal
// ========

// extractLoggerData extracts and formats given neo4j.Result, error and source string to logger message and logrus.Fields.
//
// Logs like:
// 	[Neo4j] Connection error: dial tcp [::1]:7687: connectex: No connection could be made because the target machine actively refused it. | F:/Projects/ahlib-db/xneo4j/xneo4j_test.go:97
// 	[Neo4j] Server error: [Neo.ClientError.Statement.SyntaxError] Invalid input 'n' (line 1, column 26 (offset: 25)) | F:/Projects/ahlib-db/xneo4j/xneo4j_test.go:97
// 	[Neo4j]     -1 |        999ms | MATCH (n {uid: 8}) RETURN n LIMIT 1 | F:/Projects/ahlib-db/xneo4j/xneo4j_test.go:97
// 	       |------| |------------| |-----------------------------------| |---------------------------------------------|
// 	          6           12                        ...                                          ...
func extractLoggerData(result neo4j.Result, err error, source string, options *loggerOptions) (string, logrus.Fields, bool) {
	var msg string
	var fields logrus.Fields
	var isErr bool

	if err != nil {
		// failed to connect (Connection error)
		// the target machine actively refused it
		// ...
		isErr = true
		msg = fmt.Sprintf("[Neo4j] %v | %s", err, source)
		fields = logrus.Fields{"module": "neo4j", "type": "connection", "error": err, "source": source}
	} else if summary, err := result.Summary(); err != nil {
		// failed to execute (Server error)
		// Neo.ClientError.Security.Unauthorized
		// Neo.ClientError.Statement.SyntaxError
		// Neo.ClientError.Statement.TypeError
		// Neo.ClientError.Schema.ConstraintValidationFailed
		// ...
		isErr = true
		msg = fmt.Sprintf("[Neo4j] %v | %s", err, source)
		fields = logrus.Fields{"module": "neo4j", "type": "server", "error": err, "source": source}
	} else {
		stat := summary.Statement()
		cypher := render(stat.Text(), stat.Params())
		du := summary.ResultAvailableAfter() + summary.ResultConsumedAfter()
		msg = fmt.Sprintf("[Neo4j] %6d | %12s | %s | %s", -1 /* <<< */, du, cypher, source)
		fields = logrus.Fields{
			"module":   "neo4j",
			"type":     "cypher",
			"cypher":   cypher,
			"rows":     -1, // unable to get rows, because this behavior will consume the iterator
			"duration": du,
			"source":   source,
		}
		if options.counterField {
			// also contains counter statistics
			counters := summary.Counters()
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
	}

	return msg, fields, isErr
}

// render renders given cypher string and parameters to form the cypher expression.
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
		// 1. Temporal values: https://neo4j.com/docs/cypher-manual/current/syntax/temporal/
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

		// 2. Spatial values: https://neo4j.com/docs/cypher-manual/3.5/syntax/spatial/
		case neo4j.Point:
			values[k] = value.String()

		// 3. Slice and map values:
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

		// 4. Other types ...
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
