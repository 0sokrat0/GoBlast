package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger // Глобальный логгер

func InitLogger(env string) error {
	var config zap.Config
	if env == "development" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	var err error
	Log, err = config.Build(zap.AddCaller(), zap.Fields(zap.String("service", "GoBlast")))
	if err != nil {
		return err
	}
	return nil
}

func SyncLogger() {
	if Log != nil {
		_ = Log.Sync()
	}
}
