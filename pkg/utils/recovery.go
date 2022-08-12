package utils

import (
	"go.uber.org/zap"
)

func Recover(logger *zap.Logger) {
	if r := recover(); r != nil {
		logger.Error("panic", zap.Any("panic", r))
	}
}
