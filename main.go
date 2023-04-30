package main

import (
	"context"
	"github.com/falmar/ecs-service-deployer/cmd/standalone"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func main() {
	var lConfig zap.Config
	lConfig = zap.NewDevelopmentConfig()
	lConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	lConfig.DisableStacktrace = true

	if os.Getenv("DEBUG") != "" {
		lConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		lConfig.DisableStacktrace = false
	}

	logger, _ := lConfig.Build()

	ctx := context.Background()
	if err := standalone.Cmd(logger).ExecuteContext(ctx); err != nil {
		logger.Error("Error executing command", zap.Error(err))
		os.Exit(1)
	}
	os.Exit(0)
}
