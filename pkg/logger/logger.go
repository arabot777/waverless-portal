package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/wavespeedai/waverless-portal/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger
var sugar *zap.SugaredLogger

func init() {
	defaultConfig := zap.NewDevelopmentConfig()
	defaultConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	defaultConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	defaultLogger, _ := defaultConfig.Build(zap.AddCallerSkip(1))
	Log = defaultLogger
	sugar = defaultLogger.Sugar()
}

func Init() error {
	cfg := config.GlobalConfig.Logger

	atomicLevel := zap.NewAtomicLevel()
	switch cfg.Level {
	case "debug":
		atomicLevel.SetLevel(zapcore.DebugLevel)
	case "info":
		atomicLevel.SetLevel(zapcore.InfoLevel)
	case "warn":
		atomicLevel.SetLevel(zapcore.WarnLevel)
	case "error":
		atomicLevel.SetLevel(zapcore.ErrorLevel)
	default:
		atomicLevel.SetLevel(zapcore.InfoLevel)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var syncer zapcore.WriteSyncer
	switch cfg.Output {
	case "file":
		syncer = zapcore.AddSync(getLogWriter(cfg))
	case "both":
		syncer = zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(getLogWriter(cfg)),
		)
	default:
		syncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		syncer,
		atomicLevel,
	)

	Log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = Log.Sugar()
	return nil
}

func getLogWriter(cfg config.LoggerConfig) zapcore.WriteSyncer {
	maxSize := cfg.File.MaxSize
	if maxSize == 0 {
		maxSize = 100
	}
	maxBackups := cfg.File.MaxBackups
	if maxBackups == 0 {
		maxBackups = 3
	}
	maxAge := cfg.File.MaxAge
	if maxAge == 0 {
		maxAge = 7
	}
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.File.Path,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   cfg.File.Compress,
	})
}

func Infof(format string, args ...interface{}) {
	sugar.Infof(format, args...)
}

func Errorf(format string, args ...interface{}) {
	sugar.Errorf(format, args...)
}

func Warnf(format string, args ...interface{}) {
	sugar.Warnf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	sugar.Debugf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	sugar.Fatalf(format, args...)
}

func InfoCtx(ctx context.Context, format string, args ...interface{}) {
	sugar.Infof(fmt.Sprintf("[%s] ", getTraceID(ctx))+format, args...)
}

func ErrorCtx(ctx context.Context, format string, args ...interface{}) {
	sugar.Errorf(fmt.Sprintf("[%s] ", getTraceID(ctx))+format, args...)
}

func WarnCtx(ctx context.Context, format string, args ...interface{}) {
	sugar.Warnf(fmt.Sprintf("[%s] ", getTraceID(ctx))+format, args...)
}

func FatalCtx(ctx context.Context, format string, args ...interface{}) {
	sugar.Fatalf(fmt.Sprintf("[%s] ", getTraceID(ctx))+format, args...)
}

func getTraceID(ctx context.Context) string {
	if ctx == nil {
		return "0"
	}
	return "0"
}

func Sync() error {
	return Log.Sync()
}
