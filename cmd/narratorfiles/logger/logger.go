package logger

import "log"

type Logger interface {
	Printf(format string, args ...interface{})
}

func New() *log.Logger {
	return log.Default()
}

type ErrorReporter struct {
	Logger Logger
}

func (e *ErrorReporter) ReportError(err error) {
	if err != nil {
		e.Logger.Printf("[ERROR]: %v", err)
	}
}
