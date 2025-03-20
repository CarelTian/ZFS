package logger

import (
	"ZFS/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"path/filepath"
)

var Log *zap.Logger

// InitLogger 根据配置文件初始化 zap 日志
func InitLogger(cfg *config.Config) {
	if !cfg.Log.Enable {
		Log = zap.NewNop()
		zap.ReplaceGlobals(Log)
		return
	}
	for _, path := range cfg.Log.OutputPaths {
		if err := EnsureLogPath(path); err != nil {
			log.Fatalf("创建日志失败: %v", err)
		}
	}
	for _, path := range cfg.Log.ErrorOutputPaths {
		if err := EnsureLogPath(path); err != nil {
			log.Fatalf("创建错误日志失败: %v", err)
		}
	}
	zapConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(parseLogLevel(cfg.Log.Level)),
		Development: cfg.Log.Level == "debug",
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         cfg.Log.Encoding,
		EncoderConfig:    encoderConfig(),
		OutputPaths:      cfg.Log.OutputPaths,
		ErrorOutputPaths: cfg.Log.ErrorOutputPaths,
	}
	var err error
	Log, err = zapConfig.Build()
	if err != nil {
		log.Fatalf("构建zap错误日志失败: %v", err)
	}

	zap.ReplaceGlobals(Log)
}

// parseLogLevel 将字符串级别转换为 zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// encoderConfig 返回日志的编码配置
func encoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // 时间格式化
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // caller 信息缩写
	}
}

func EnsureLogPath(path string) error {
	if path == "stdout" || path == "stderr" {
		return nil
	}
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0777) //为了方便，所有用户都是最高权限
	}
	return nil
}
