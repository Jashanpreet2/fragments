package logger

import (
	"go.uber.org/zap"
)

var Sugar *zap.SugaredLogger

func init() {
	// Logger setup
	// logger = getLogger(os.Getenv("LOG_LEVEL"))
	logger, _ := zap.NewDevelopment()

	Sugar = logger.Sugar()
}
