package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
)

// InitLogger return a zap logger
func InitLogger(debug bool, logPath string) *zap.Logger {
	cores := []zapcore.Core{}
	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		TimeKey:     "time",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	})
	consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		TimeKey:     "time",
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	})

	if debug {
		cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zap.DebugLevel))
	}
	cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zap.InfoLevel))
	cores = append(cores, zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stderr), zap.ErrorLevel))

	if logPath != "" {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Join(logPath, "recorder.log"),
			MaxSize:    5, // megabytes
			MaxBackups: 5,
			MaxAge:     60, // days
		})

		if debug {
			cores = append(cores, zapcore.NewCore(jsonEncoder, fileWriter, zap.DebugLevel))
		}
		cores = append(cores, zapcore.NewCore(jsonEncoder, fileWriter, zap.InfoLevel))
		cores = append(cores, zapcore.NewCore(jsonEncoder, fileWriter, zap.ErrorLevel))
	}

	core := zapcore.NewTee(cores...)

	return zap.New(core)
}
