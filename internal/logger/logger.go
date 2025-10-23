package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

var (
	log *slog.Logger
)

func Init(level, output string) error {
	slogLevel := parseLevel(level)
	addSource := slogLevel == slog.LevelDebug

	var writer io.Writer
	var useColor bool

	switch output {
	case "stdout", "":
		writer = os.Stdout
		useColor = isTerminal(os.Stdout)
	case "stderr":
		writer = os.Stderr
		useColor = isTerminal(os.Stderr)
	default:
		file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writer = file
		useColor = false
	}

	var handler slog.Handler
	if useColor {
		handler = newColorHandler(writer, &slog.HandlerOptions{
			Level:     slogLevel,
			AddSource: addSource,
		})
	} else {
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level:     slogLevel,
			AddSource: addSource,
		})
	}

	log = slog.New(handler)
	return nil
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func isTerminal(f *os.File) bool {
	fileInfo, err := f.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func Sync() {
}

func Debug(args ...interface{}) {
	log.Debug(sprint(args...))
}

func Debugf(template string, args ...interface{}) {
	log.Debug(sprintf(template, args...))
}

func Info(args ...interface{}) {
	log.Info(sprint(args...))
}

func Infof(template string, args ...interface{}) {
	log.Info(sprintf(template, args...))
}

func Warn(args ...interface{}) {
	log.Warn(sprint(args...))
}

func Warnf(template string, args ...interface{}) {
	log.Warn(sprintf(template, args...))
}

func Error(args ...interface{}) {
	log.Error(sprint(args...))
}

func Errorf(template string, args ...interface{}) {
	log.Error(sprintf(template, args...))
}

func Fatal(args ...interface{}) {
	log.Error(sprint(args...))
	os.Exit(1)
}

func Fatalf(template string, args ...interface{}) {
	log.Error(sprintf(template, args...))
	os.Exit(1)
}

func With(args ...interface{}) *slog.Logger {
	return log.With(args...)
}

func Logger() *slog.Logger {
	return log
}

func sprint(args ...interface{}) string {
	return fmt.Sprint(args...)
}

func sprintf(template string, args ...interface{}) string {
	return fmt.Sprintf(template, args...)
}
