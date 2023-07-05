package internal

import (
	"os"

	"go.uber.org/zap"
)

// NewLogger returns a new logger.
func NewLogger() (*zap.SugaredLogger, error) {
	var logger *zap.Logger
	var err error

	if os.Getenv("ENVIRONMENT") == "development" {
		logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
	} else {
		logger, err = zap.NewProduction()
		if err != nil {
			return nil, err
		}
	}

	return logger.Sugar(), nil
}
