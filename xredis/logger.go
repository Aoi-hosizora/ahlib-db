package xredis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"log"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// logrus

type LogrusLogger struct {
	redis.Conn
	logger  *logrus.Logger
	LogMode bool
	Skip    int
}

func NewLogrusLogger(conn redis.Conn, logger *logrus.Logger, logMode bool) *LogrusLogger {
	return &LogrusLogger{Conn: conn, logger: logger, LogMode: logMode, Skip: 2}
}

func (l *LogrusLogger) WithSkip(skip int) *LogrusLogger {
	l.Skip = skip
	return l
}

func (l *LogrusLogger) Do(commandName string, args ...interface{}) (interface{}, error) {
	s := time.Now()
	reply, err := l.Conn.Do(commandName, args...)
	e := time.Now()

	if l.LogMode {
		l.print(reply, err, commandName, e.Sub(s).String(), args...)
	}

	return reply, err
}

func (l *LogrusLogger) print(reply interface{}, err error, commandName string, du string, v ...interface{}) {
	cmd := renderCommand(commandName, v)
	_, file, line, _ := runtime.Caller(l.Skip)
	source := fmt.Sprintf("%s:%d", file, line)

	if err != nil {
		l.logger.WithFields(logrus.Fields{
			"module":  "redis",
			"command": cmd,
			"error":   err,
			"source":  source,
		}).Error(fmt.Sprintf("[Redis] %v | %s | %s", err, cmd, source))
		return
	}
	if cmd == "" {
		return
	}

	cnt, t := renderReply(reply)
	l.logger.WithFields(logrus.Fields{
		"module":   "redis",
		"command":  cmd,
		"count":    cnt,
		"duration": du,
		"source":   source,
	}).Info(fmt.Sprintf("[Redis] #: %3d | %12s | %15s | %s | %s", cnt, du, t, cmd, source))
}

// logger

type LoggerRedis struct {
	redis.Conn
	logger  *log.Logger
	LogMode bool
	Skip    int
}

func NewLoggerRedis(conn redis.Conn, logger *log.Logger, logMode bool) *LoggerRedis {
	return &LoggerRedis{Conn: conn, logger: logger, LogMode: logMode, Skip: 2}
}

func (l *LoggerRedis) WithSkip(skip int) *LoggerRedis {
	l.Skip = skip
	return l
}

func (l *LoggerRedis) Do(commandName string, args ...interface{}) (interface{}, error) {
	s := time.Now()
	reply, err := l.Conn.Do(commandName, args...)
	e := time.Now()

	if l.LogMode {
		l.print(reply, err, commandName, e.Sub(s).String(), args...)
	}

	return reply, err
}

func (l *LoggerRedis) print(reply interface{}, err error, commandName string, du string, v ...interface{}) {
	cmd := renderCommand(commandName, v)
	_, file, line, _ := runtime.Caller(l.Skip)
	source := fmt.Sprintf("%s:%d", file, line)

	if err != nil {
		l.logger.Printf("[Redis] %v | %s | %s", err, cmd, source)
		return
	}
	if cmd == "" {
		return
	}

	cnt, t := renderReply(reply)
	l.logger.Printf("[Redis] #: %3d | %12s | %15s | %s | %s", cnt, du, t, cmd, source)
}

// render

func renderCommand(cmd string, args []interface{}) string {
	out := cmd
	for _, arg := range args {
		out += " " + fmt.Sprintf("%v", arg)
	}
	return strings.TrimSpace(out)
}

func renderReply(reply interface{}) (cnt int, t string) {
	if reply == nil {
		cnt = 0
		t = "<nil>"
	} else if val := reflect.ValueOf(reply); val.Kind() == reflect.Slice && val.IsValid() {
		cnt = val.Len()
		t = val.Type().Elem().String()
		if t == "uint8" { // byte
			cnt = 1
			t = "string"
		} else if t == "interface {}" && val.Len() >= 1 {
			t = reflect.TypeOf(val.Index(0).Interface()).String()
			if t == "[]uint8" { // string
				t = "string"
			}
		}
	} else {
		cnt = 1
		t = fmt.Sprintf("%T", reply)
	}
	if reply == "OK" {
		t = "string (OK)"
	}
	return
}
