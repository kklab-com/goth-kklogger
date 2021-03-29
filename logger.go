package kklogger

type Logger interface {
	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type LoggerHook interface {
	Logger
}

type DefaultLoggerHook struct {
}

func (h *DefaultLoggerHook) Trace(args ...interface{}) {
}

func (h *DefaultLoggerHook) Debug(args ...interface{}) {
}

func (h *DefaultLoggerHook) Info(args ...interface{}) {
}

func (h *DefaultLoggerHook) Warn(args ...interface{}) {
}

func (h *DefaultLoggerHook) Error(args ...interface{}) {
}
