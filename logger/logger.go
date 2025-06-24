package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Log 全局日志对象
	Log *zap.Logger
)

// Config 日志配置
type Config struct {
	Level      string // 日志级别
	Filename   string // 日志文件路径
	MaxSize    int    // 单个文件最大尺寸，单位MB
	MaxBackups int    // 最大保留文件数
	MaxAge     int    // 最大保留天数
	Compress   bool   // 是否压缩
}

// Init 初始化日志
func Init(config *Config) error {
	// 创建日志目录
	if err := os.MkdirAll(filepath.Dir(config.Filename), 0755); err != nil {
		return fmt.Errorf("create log directory failed: %v", err)
	}

	// 设置日志级别
	level := zap.InfoLevel
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}

	// 配置日志轮转
	writer := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(writer)),
		level,
	)

	// 创建日志对象
	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return nil
}

// Debug 输出调试日志
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Info 输出信息日志
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Warn 输出警告日志
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Error 输出错误日志
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Fatal 输出致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// With 创建子日志对象
func With(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

// Field 创建日志字段
func Field(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// String 创建字符串字段
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int 创建整数字段
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Float64 创建浮点数字段
func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

// Time 创建时间字段
func Time(key string, value time.Time) zap.Field {
	return zap.Time(key, value)
}

// Duration 创建时间间隔字段
func Duration(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

// Infof 格式化输出信息日志
func Infof(format string, args ...interface{}) {
	Log.Info(fmt.Sprintf(format, args...))
}

// Errorf 格式化输出错误日志
func Errorf(format string, args ...interface{}) {
	Log.Error(fmt.Sprintf(format, args...))
}

// Warnf 格式化输出警告日志
func Warnf(format string, args ...interface{}) {
	Log.Warn(fmt.Sprintf(format, args...))
}

// Debugf 格式化输出调试日志
func Debugf(format string, args ...interface{}) {
	Log.Debug(fmt.Sprintf(format, args...))
}

// Fatalf 格式化输出致命错误日志
func Fatalf(format string, args ...interface{}) {
	Log.Fatal(fmt.Sprintf(format, args...))
}
