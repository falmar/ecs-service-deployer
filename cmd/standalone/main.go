package main

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func initConfig(l *zap.Logger) func() {
	return func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")

		if err := viper.ReadInConfig(); err != nil {
			l.Fatal("Error reading config file", zap.Error(err))
		}
	}
}

func main() {
	config := zap.NewDevelopmentConfig()
	config.DisableCaller = true
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ := config.Build()

	cobra.OnInitialize(initConfig(logger))
	ctx := context.Background()

	var rootCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deployer is a lambda function that deploys a ECS Service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.Error("Error executing command", zap.Error(err))
		os.Exit(1)
	}
	os.Exit(0)
}
