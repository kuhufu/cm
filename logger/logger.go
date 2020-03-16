package log

import (
	"log"
	"sync/atomic"
)

type Level uint32

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

var strMap = []string{
	"panic",
	"fatal",
	"error",
	"warn",
	"info",
	"debug",
	"trace",
}

func (l Level) String() string {
	return strMap[l]
}

type Logger struct {
	Level Level
}

func (logger *Logger) level() Level {
	return Level(atomic.LoadUint32((*uint32)(&logger.Level)))
}

func (logger *Logger) setLevel(level Level) {
	atomic.StoreUint32((*uint32)(&logger.Level), uint32(level))
}

func (logger *Logger) IsLevelEnabled(level Level) bool {
	return logger.level() >= level
}

func (l *Logger) Logf(level Level, format string, args ...interface{}) {
	if l.IsLevelEnabled(level) {
		log.Printf(format, args...)
	}
}

func (l *Logger) Log(level Level, args ...interface{}) {
	if l.IsLevelEnabled(level) {
		log.Print(args...)
	}
}

func (l *Logger) Logln(level Level, args ...interface{}) {
	if l.IsLevelEnabled(level) {
		log.Println(args...)
	}
}

func (l *Logger) Trace(args ...interface{}) {
	l.Log(TraceLevel, args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.Log(DebugLevel, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.Log(InfoLevel, args...)
}

func (l *Logger) Print(args ...interface{}) {
	l.Log(InfoLevel, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.Log(WarnLevel, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.Log(ErrorLevel, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.Log(FatalLevel, args...)
}

func (l *Logger) Panic(args ...interface{}) {
	l.Log(PanicLevel, args...)
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	l.Logf(TraceLevel, format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logf(DebugLevel, format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logf(InfoLevel, format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logf(WarnLevel, format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logf(ErrorLevel, format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.Logf(FatalLevel, format, args...)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.Logf(PanicLevel, format, args...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.Logf(InfoLevel, format, args...)
}

func (l *Logger) Println(args ...interface{}) {
	l.Logln(InfoLevel, args...)
}
