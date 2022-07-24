package xneo4j

import (
	"encoding/json"
	"fmt"
	"github.com/Aoi-hosizora/ahlib/xcolor"
	"github.com/Aoi-hosizora/ahlib/xstring"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/sirupsen/logrus"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

// loggerOptions is a type of LogrusLogger and StdLogger's option, each field can be set by LoggerOption function type.
type loggerOptions struct {
	logErr        bool
	logCypher     bool
	counterFields bool
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

// WithLogCypher creates a LoggerOption to decide whether to do log for cyphers or not, defaults to true.
func WithLogCypher(log bool) LoggerOption {
	return func(o *loggerOptions) {
		o.logCypher = log
	}
}

// WithCounterFields creates a LoggerOption to decide whether to do store counter fields to logrus.Fields or not, defaults to false.
func WithCounterFields(flag bool) LoggerOption {
	return func(o *loggerOptions) {
		o.counterFields = flag
	}
}

// WithSkip creates a LoggerOption to specify runtime skip for getting runtime information, defaults to 1.
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
	opt := &loggerOptions{logErr: true, logCypher: true, counterFields: false, skip: 1}
	for _, o := range options {
		if o != nil {
			o(opt)
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
	opt := &loggerOptions{logErr: true, logCypher: true, counterFields: false, skip: 1}
	for _, o := range options {
		if o != nil {
			o(opt)
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
	enable := _enable.Load().(bool)
	if !enable {
		return result, err // ignore
	}

	_, file, line, _ := runtime.Caller(l.option.skip) // defaults to 4
	source := fmt.Sprintf("%s:%d", file, line)

	p := extractLoggerParam(result, err, source, l.option)
	m := formatLoggerParam(p)
	f := fieldifyLoggerParam(p)
	if p.ErrorMsg != "" && l.option.logErr {
		l.logger.WithFields(f).Error(m)
	} else if p.ErrorMsg == "" && l.option.logCypher {
		l.logger.WithFields(f).Info(m)
	}
	return result, err
}

// Run implements neo4j.Session interface, it executes given cypher and params, and logs neo4j's message to logrus.StdLogger.
func (l *StdLogger) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error) {
	result, err := l.Session.Run(cypher, params, configurers...)
	enable := _enable.Load().(bool)
	if !enable {
		return result, err // ignore
	}

	_, file, line, _ := runtime.Caller(l.option.skip) // defaults to 4
	source := fmt.Sprintf("%s:%d", file, line)

	p := extractLoggerParam(result, err, source, l.option)
	m := formatLoggerParam(p)
	if (p.ErrorMsg != "" && l.option.logErr) || (p.ErrorMsg == "" && l.option.logCypher) {
		l.logger.Print(m)
	}
	return result, err
}

// =======
// extract
// =======

// LoggerParam stores some logger parameters and is used by LogrusLogger and StdLogger.
type LoggerParam struct {
	Type     string
	Cypher   string
	Duration time.Duration
	Slow     bool
	Source   string
	Counter  neo4j.Counters
	ErrorMsg string
}

var (
	// FormatLoggerFunc is a custom LoggerParam's format function for LogrusLogger and StdLogger.
	FormatLoggerFunc func(p *LoggerParam) string

	// FieldifyLoggerFunc is a custom LoggerParam's fieldify function for LogrusLogger.
	FieldifyLoggerFunc func(p *LoggerParam) logrus.Fields
)

// extractLoggerParam extracts and returns LoggerParam using given parameters.
func extractLoggerParam(result neo4j.Result, err error, source string, options *loggerOptions) *LoggerParam {
	if err != nil {
		// 1. failed to connect (Connection error)
		// the target machine actively refused it
		// ...
		return &LoggerParam{Type: "connection", ErrorMsg: err.Error(), Source: source}
	}

	summary, err := result.Summary()
	if err != nil {
		// 2. failed to execute (Server error)
		// Neo.ClientError.Security.Unauthorized
		// Neo.ClientError.Statement.SyntaxError
		// Neo.ClientError.Statement.TypeError
		// Neo.ClientError.Schema.ConstraintValidationFailed
		// ...
		lines := make([]string, 0, 2)
		for _, m := range strings.Split(strings.ReplaceAll(err.Error(), "^", ""), "\n") {
			m = strings.TrimSpace(xstring.RemoveExtraBlanks(m))
			if len(m) > 0 {
				lines = append(lines, m)
			}
		}
		errMsg := strings.Join(lines, "; ")
		return &LoggerParam{Type: "Server", ErrorMsg: errMsg, Source: source}
	}

	stat := summary.Statement()
	cypher := render(stat.Text(), stat.Params())
	du := summary.ResultAvailableAfter() + summary.ResultConsumedAfter()
	slow := options.slowThreshold > 0 && du >= options.slowThreshold
	p := &LoggerParam{
		Type:     "Cypher",
		Cypher:   cypher,
		Duration: du,
		Slow:     slow,
		Source:   source,
	}
	if options.counterFields {
		p.Counter = summary.Counters()
	}
	return p
}

// formatLoggerParam formats given LoggerParam to string for LogrusLogger and StdLogger.
//
// The default format logs like:
// 	[Neo4j] err: Connection error: dial tcp [::1]:7687: connectex: No connection could be made because the target machine actively refused it. | F:/Projects/ahlib-db/xneo4j/xneo4j_test.go:97
// 	[Neo4j] err: Server error: [Neo.ClientError.Statement.SyntaxError] Invalid input 'n' (line 1, column 26 (offset: 25)) | F:/Projects/ahlib-db/xneo4j/xneo4j_test.go:97
// 	[Neo4j]     -1 |        999ms | MATCH (n {uid: 8}) RETURN n LIMIT 1 | F:/Projects/ahlib-db/xneo4j/xneo4j_test.go:97
// 	       |------| |------------| |-----------------------------------| |---------------------------------------------|
// 	          6           12                        ...                                          ...
func formatLoggerParam(p *LoggerParam) string {
	if FormatLoggerFunc != nil {
		return FormatLoggerFunc(p)
	}
	if p.ErrorMsg != "" {
		return fmt.Sprintf("[Neo4j] err: %v | %s", p.ErrorMsg, p.Source)
	}
	du := fmt.Sprintf("%12s", p.Duration.String())
	if p.Slow {
		du = xcolor.Yellow.WithStyle(xcolor.Bold).Sprintf("%12s", p.Duration.String())
	}
	return fmt.Sprintf("[Neo4j] %6d | %s | %s | %s", -1, du, p.Cypher, p.Source)
}

// fieldifyLoggerParam fieldifies given LoggerParam to logrus.Fields for LogrusLogger.
//
// The default contains the following fields: module, type, cypher, duration, source, error, and counter fields.
func fieldifyLoggerParam(p *LoggerParam) logrus.Fields {
	if FieldifyLoggerFunc != nil {
		return FieldifyLoggerFunc(p)
	}
	if p.ErrorMsg != "" {
		return logrus.Fields{
			"module": "neo4j",
			"type":   p.Type, // connection / server
			"error":  p.ErrorMsg,
			"source": p.Source,
		}
	}
	f := logrus.Fields{
		"module":   "neo4j",
		"type":     p.Type, // cypher
		"cypher":   p.Cypher,
		"duration": p.Duration,
		"source":   p.Source,
	}
	if p.Counter != nil {
		f["nodesCreated"] = p.Counter.NodesCreated()
		f["nodesDeleted"] = p.Counter.NodesDeleted()
		f["relationshipsCreated"] = p.Counter.RelationshipsCreated()
		f["relationshipsDeleted"] = p.Counter.RelationshipsDeleted()
		f["propertiesSet"] = p.Counter.PropertiesSet()
		f["labelsAdded"] = p.Counter.LabelsAdded()
		f["labelsRemoved"] = p.Counter.LabelsRemoved()
		f["indexesAdded"] = p.Counter.IndexesAdded()
		f["indexesRemoved"] = p.Counter.IndexesRemoved()
		f["constraintsAdded"] = p.Counter.ConstraintsAdded()
		f["constraintsRemoved"] = p.Counter.ConstraintsRemoved()
	}
	return f
}

// ========
// internal
// ========

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
		placeholder := fmt.Sprintf(`\$%s([^\w]|$)`, k) // `\$...([^\w]|$)` or `\$...(\W)`
		result = regexp.MustCompile(placeholder).ReplaceAllString(result, v+"$1")
	}
	return strings.TrimSpace(result)
}
