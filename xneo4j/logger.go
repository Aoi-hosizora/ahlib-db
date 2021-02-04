package xneo4j

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/sirupsen/logrus"
	"reflect"
	"runtime"
	"strings"
)

// loggerOptions represents some options for logger, set by LoggerOption.
type loggerOptions struct {
	skip int // runtime skip
}

// LoggerOption represents an option for logger, created by WithXXX functions.
type LoggerOption func(*loggerOptions)

// WithSkip returns a LoggerOption with runtime skip to get runtime information, defaults to 1.
func WithSkip(skip int) LoggerOption {
	return func(o *loggerOptions) {
		o.skip = skip
	}
}

// LogrusLogger logs neo4j cypher executing message to logrus.Logger.
type LogrusLogger struct {
	neo4j.Session
	logger  *logrus.Logger
	options *loggerOptions
}

var _ neo4j.Session = &LogrusLogger{}

// NewLogrusLogger creates a new LogrusLogger using given logrus.Logger and LoggerOption-s.
func NewLogrusLogger(session neo4j.Session, logger *logrus.Logger, options ...LoggerOption) *LogrusLogger {
	opt := &loggerOptions{
		skip: 1, // default to 1
	}
	for _, op := range options {
		if op != nil {
			op(opt)
		}
	}
	return &LogrusLogger{Session: session, logger: logger, options: opt}
}

// LoggerLogger logs neo4j cypher executing message to logrus.StdLogger.
type LoggerLogger struct {
	neo4j.Session
	logger  logrus.StdLogger
	options *loggerOptions
}

var _ neo4j.Session = &LoggerLogger{}

// NewLoggerLogger creates a new LoggerLogger using given log.Logger and LoggerOption-s.
func NewLoggerLogger(session neo4j.Session, logger logrus.StdLogger, options ...LoggerOption) *LoggerLogger {
	opt := &loggerOptions{
		skip: 1, // default to 1
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

	_, file, line, _ := runtime.Caller(l.options.skip)
	source := fmt.Sprintf("%s:%d", file, line)
	msg, fields, isErr := formatLoggerAndFields(result, err, source)
	if isErr {
		l.logger.WithFields(fields).Error(msg)
	} else {
		l.logger.WithFields(fields).Info(msg)
	}

	return result, err
}

// Run executes given cypher and params, and logs to logrus.StdLogger.
func (l *LoggerLogger) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error) {
	result, err := l.Session.Run(cypher, params, configurers...)

	_, file, line, _ := runtime.Caller(l.options.skip)
	source := fmt.Sprintf("%s:%d", file, line)
	msg, _, _ := formatLoggerAndFields(result, err, source)
	l.logger.Print(msg)

	return result, err
}

// formatLoggerAndFields formats given neo4j.Result, error and source to logger string, logrus.Fields and isError flag.
// Logs like:
// 	[Neo4j] xxx
func formatLoggerAndFields(result neo4j.Result, err error, source string) (string, logrus.Fields, bool) {
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
		// Neo.ClientError.Statement.SyntaxError
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
			"rows":     0, // unable to get rows, because this behavior will consume the iterator
			"duration": du,
			"source":   source,
			// some stat
			"nodesCreated":         counters.NodesCreated(),
			"nodesDeleted":         counters.NodesDeleted(),
			"relationshipsCreated": counters.RelationshipsCreated(),
			"relationshipsDeleted": counters.RelationshipsDeleted(),
			"propertiesSet":        counters.PropertiesSet(),
			"labelsAdded":          counters.LabelsAdded(),
			"labelsRemoved":        counters.LabelsRemoved(),
			"indexesAdded":         counters.IndexesAdded(),
			"indexesRemoved":       counters.IndexesRemoved(),
			"constraintsAdded":     counters.ConstraintsAdded(),
			"constraintsRemoved":   counters.ConstraintsRemoved(),
		}
		msg = fmt.Sprintf("[Neo4j] #: %3d | %12s | %s | %s", -1, du, cypher, source)
	}

	return msg, fields, isErr
}

// render renders cypher string and parameters to complete cypher expression.
func render(cypher string, params map[string]interface{}) string {
	out := cypher
	for k, v := range params {
		to := fmt.Sprintf("%v", v)
		if reflect.TypeOf(v).Kind() == reflect.String {
			to = "'" + to + "'"
		}
		out = strings.ReplaceAll(out, "$"+k, to)
	}
	return out
}
