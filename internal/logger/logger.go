// Package logger 提供简化的日志功能
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log *zap.SugaredLogger
)

// Init 初始化日志系统
func Init(level, output string) error {
	// 解析日志级别
	zapLevel, err := parseLevel(level)
	if err != nil {
		zapLevel = zapcore.InfoLevel
	}

	// 配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.ConsoleSeparator = "\t"  // 使用 Tab 分隔字段

	// 配置输出
	var writer zapcore.WriteSyncer
	switch output {
	case "stdout", "":
		writer = zapcore.AddSync(os.Stdout)
	case "stderr":
		writer = zapcore.AddSync(os.Stderr)
	default:
		// 文件输出
		file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writer = zapcore.AddSync(file)
	}

	// 创建核心
	core := zapcore.NewCore(
		newAlignedEncoder(encoderConfig),
		writer,
		zapLevel,
	)

	// 创建 logger
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	log = logger.Sugar()

	return nil
}

// parseLevel 解析日志级别
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}

// Sync 刷新日志缓冲
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

// Debug 输出 debug 级别日志
func Debug(args ...interface{}) {
	log.Debug(args...)
}

// Debugf 格式化输出 debug 级别日志
func Debugf(template string, args ...interface{}) {
	log.Debugf(template, args...)
}

// Info 输出 info 级别日志
func Info(args ...interface{}) {
	log.Info(args...)
}

// Infof 格式化输出 info 级别日志
func Infof(template string, args ...interface{}) {
	log.Infof(template, args...)
}

// Warn 输出 warn 级别日志
func Warn(args ...interface{}) {
	log.Warn(args...)
}

// Warnf 格式化输出 warn 级别日志
func Warnf(template string, args ...interface{}) {
	log.Warnf(template, args...)
}

// Error 输出 error 级别日志
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf 格式化输出 error 级别日志
func Errorf(template string, args ...interface{}) {
	log.Errorf(template, args...)
}

// Fatal 输出 fatal 级别日志并退出程序
func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

// Fatalf 格式化输出 fatal 级别日志并退出程序
func Fatalf(template string, args ...interface{}) {
	log.Fatalf(template, args...)
}

// With 添加结构化字段
func With(args ...interface{}) *zap.SugaredLogger {
	return log.With(args...)
}
