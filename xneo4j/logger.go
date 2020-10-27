package xneo4j

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"github.com/sirupsen/logrus"
	"log"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// logrus

type LogrusNeo4j struct {
	neo4j.Session
	logger  *logrus.Logger
	LogMode bool
	Skip    int
}

func NewLogrusNeo4j(session neo4j.Session, logger *logrus.Logger, logMode bool) *LogrusNeo4j {
	return &LogrusNeo4j{Session: session, logger: logger, LogMode: logMode, Skip: 2}
}

func (l *LogrusNeo4j) WithSkip(skip int) *LogrusNeo4j {
	l.Skip = skip
	return l
}

func (l *LogrusNeo4j) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error) {
	s := time.Now()
	result, err := l.Session.Run(cypher, params, configurers...)
	e := time.Now()

	if l.LogMode {
		l.print(result, e.Sub(s).String(), err)
	}

	return result, err
}

func (l *LogrusNeo4j) print(result neo4j.Result, du string, err error) {
	_, file, line, _ := runtime.Caller(l.Skip)
	source := fmt.Sprintf("%s:%d", file, line)

	if err != nil { // Failed to run cypher.
		l.logger.WithFields(logrus.Fields{
			"module": "neo4j",
			"error":  err,
			"source": source,
		}).Error(fmt.Sprintf("[Neo4j] error: %v | %s", err, source))
		return
	}

	summary, err := result.Summary()
	if err != nil { // Failed to get summary.
		// Neo.ClientError.Statement.SyntaxError
		// Neo.ClientError.Schema.ConstraintValidationFailed
		// ...
		l.logger.WithFields(logrus.Fields{
			"module": "neo4j",
			"error":  err,
			"source": source,
		}).Error(fmt.Sprintf("[Neo4j] error: %v | %s", err, source))
		return
	}

	keys, err := result.Keys()
	if err != nil { // Failed to get keys.
		l.logger.WithFields(logrus.Fields{
			"module": "neo4j",
			"error":  err,
			"source": source,
		}).Error(fmt.Sprintf("[Neo4j] error: %v | %s", err, source))
		return
	}

	// Success to run cypher and get summary, get information form summary.
	stat := summary.Statement()
	cypher := stat.Text()
	params := stat.Params()
	cypher = render(cypher, params)

	l.logger.WithFields(logrus.Fields{
		"module":   "neo4j",
		"cypher":   cypher,
		"rows":     0,
		"columns":  len(keys),
		"duration": du,
		"source":   source,
	}).Info(fmt.Sprintf("[Neo4j] #C: %2d | %12s | %s | %s", len(keys), du, cypher, source))
}

// logger

type LoggerNeo4j struct {
	neo4j.Session
	logger  *log.Logger
	LogMode bool
	Skip    int
}

func NewLoggerNeo4j(session neo4j.Session, logger *log.Logger, logMode bool) *LoggerNeo4j {
	return &LoggerNeo4j{Session: session, logger: logger, LogMode: logMode, Skip: 2}
}

func (l *LoggerNeo4j) WithSkip(skip int) *LoggerNeo4j {
	l.Skip = skip
	return l
}

func (l *LoggerNeo4j) Run(cypher string, params map[string]interface{}, configurers ...func(*neo4j.TransactionConfig)) (neo4j.Result, error) {
	s := time.Now()
	result, err := l.Session.Run(cypher, params, configurers...)
	e := time.Now()

	if l.LogMode {
		l.print(result, e.Sub(s).String(), err)
	}

	return result, err
}

func (l *LoggerNeo4j) print(result neo4j.Result, du string, err error) {
	_, file, line, _ := runtime.Caller(l.Skip)
	source := fmt.Sprintf("%s:%d", file, line)

	if err != nil {
		l.logger.Printf("[Neo4j] error: %v | %s", err, source)
		return
	}
	summary, err := result.Summary()
	if err != nil {
		l.logger.Printf("[Neo4j] error: %v | %s", err, source)
		return
	}
	keys, err := result.Keys()
	if err != nil {
		l.logger.Printf("[Neo4j] error: %v | %s", err, source)
		return
	}

	stat := summary.Statement()
	cypher := stat.Text()
	params := stat.Params()
	cypher = render(cypher, params)

	l.logger.Printf("[Neo4j] #C: %2d | %12s | %s | %s", len(keys), du, cypher, source)
}

// render

func render(cypher string, params map[string]interface{}) string {
	out := cypher
	for k, v := range params {
		t := reflect.TypeOf(v)
		to := fmt.Sprintf("%v", v)
		if t.Kind() == reflect.String {
			to = "'" + to + "'"
		}
		out = strings.ReplaceAll(out, "$"+k, to)
	}
	return out
}
