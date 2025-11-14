package polymarketrealtime

import "fmt"

type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Warn(format string, args ...any)
	Error(format string, args ...any)
}

type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

type logger struct {
	level LogLevel
}

func NewLogger(level LogLevel) Logger {
	return &logger{level: level}
}

func (l logger) Debug(format string, args ...any) {
	if l.level <= LogLevelDebug {
		fmt.Printf("[PolymarketRealTimeDataClient][DEBUG] "+format+"\n", args...)
	}
}

func (l logger) Info(format string, args ...any) {
	if l.level <= LogLevelInfo {
		fmt.Printf("[PolymarketRealTimeDataClient][INFO] "+format+"\n", args...)
	}
}

func (l logger) Warn(format string, args ...any) {
	if l.level <= LogLevelWarn {
		fmt.Printf("[PolymarketRealTimeDataClient][WARN] "+format+"\n", args...)
	}
}

func (l logger) Error(format string, args ...any) {
	if l.level <= LogLevelWarn {
		fmt.Printf("[PolymarketRealTimeDataClient][ERROR] "+format+"\n", args...)
	}
}

type silentLogger struct {
}

func (silentLogger) Debug(string, ...any) {}

func (silentLogger) Info(string, ...any) {}

func (silentLogger) Warn(string, ...any) {}

func (silentLogger) Error(string, ...any) {}

func NewSilentLogger() Logger {
	return silentLogger{}
}
