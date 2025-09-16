package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func epochMilliTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	//enc.AppendInt64(t.UnixMilli())
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func millisDurationEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendFloat64(float64(d) / float64(time.Millisecond))
}

func ZlogInit() *zap.Logger {
	lvl := zap.InfoLevel
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(lvl),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     epochMilliTimeEncoder,
			EncodeDuration: millisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger
}
