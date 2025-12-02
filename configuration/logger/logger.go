package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log *zap.Logger
)

func init() {
	loadconfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:   "message",
			LevelKey:     "level",
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeLevel:  zapcore.LowercaseColorLevelEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	logNovo, err := loadconfig.Build()

	if err != nil {
		fmt.Println("error iniciar logger ", err.Error())
		panic("Error ao iniciar o logger")
	}

	log = logNovo
}

func Info(message string, tags ...zap.Field) {
	log.Info(message, tags...)
}

func Error(message string, err error, tags ...zap.Field) {
	tags = append(tags, zap.NamedError("error", err))
	log.Error(message, tags...)
}

func GetLogger() *zap.Logger {
	return log
}
