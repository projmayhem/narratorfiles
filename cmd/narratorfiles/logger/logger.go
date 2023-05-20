package logger

import "log"

type Logger interface {
	Printf(format string, args ...interface{})
}

func New(l Logger) *log.Logger {
	return log.Default()
}
