package logger

import (
	"os"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func InitLogger() {
	logger := zap.Must(zap.NewDevelopment())
  if os.Getenv("APP_ENV") == "production" {
    logger = zap.Must(zap.NewProduction())
  }
	Log = logger.Sugar()
}

func SyncLogger() {
	if Log != nil {
		_ = Log.Sync()
	}
}
