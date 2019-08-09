package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志对象
var Logger *zap.Logger

// InitLogger 初始化
func InitLogger() *zap.Logger {
	Logger = initLogger("logs/proxy-b.log", "info")
	return Logger
}

// initLogger 初始化 zap 的日志
func initLogger(logpath string, loglevel string) *zap.Logger {
	hook := lumberjack.Logger{
		Filename:   logpath, // ⽇志⽂件路径
		MaxSize:    500,     // megabytes MB
		MaxBackups: 1,       // 最多保留1个备份
	}

	w := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(os.Stderr), zapcore.AddSync(&hook))

	var level zapcore.Level
	switch loglevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	// encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	// encoderConfig.CallerKey = "caller"

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		w,
		level,
	)

	logger := zap.New(core, zap.Development(), zap.AddCaller())
	return logger
}
