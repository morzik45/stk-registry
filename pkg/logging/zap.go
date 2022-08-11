package logging

import (
	"context"
	"fmt"
	"github.com/morzik45/stk-registry/pkg/config"
	"github.com/natefinch/lumberjack"
	"github.com/strpc/zaptelegram"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func NewLogger(config *config.Config) (logger *zap.Logger, err error) {
	writeSyncer := getLogWriter(config.Logger.Path)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = SyslogTimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)
	logger = zap.New(core)

	if config.Logger.Telegram.Enabled {
		logger, err = initTelegramHook(logger, config)
		if err != nil {
			return nil, err
		}
	}
	return logger, nil
}

func initTelegramHook(logger *zap.Logger, config *config.Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(config.Logger.Telegram.Level)
	if err != nil {
		level = zapcore.WarnLevel
	}

	telegramHook, err := zaptelegram.NewTelegramHook(
		config.Logger.Telegram.Token,
		config.Logger.Telegram.UsersIDs,
		zaptelegram.WithLevel(level),
		zaptelegram.WithTimeout(time.Second*3),
		zaptelegram.WithQueue(context.Background(), time.Second*3, 500),
		zaptelegram.WithFormatter(func(e zapcore.Entry) string {
			return fmt.Sprintf(
				"service: stk-registry\nlogger: %s\n%s - %s\n%s",
				e.LoggerName, e.Time.UTC().Format("2006-01-02 15:04:05"), e.Level, e.Message,
			)
		}),
	)
	if err != nil {
		return nil, err
	}
	logger = logger.WithOptions(zap.Hooks(telegramHook.GetHook()))
	return logger, nil
}

func SyslogTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func getLogWriter(logPath string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // megabytes
		MaxBackups: 10,
		MaxAge:     50, //days
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}
