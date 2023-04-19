package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
)

type Event struct {
	Image          string `json:"image"`
	Service        string `json:"service"`
	TaskDefinition string `json:"task_definition"`
}

func main() {
	//config := zap.NewDevelopmentConfig()
	//config.DisableCaller = true
	//config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	//logger, _ := config.Build()

	lambda.Start(func(ctx context.Context, e Event) (string, error) {
		if e.Image == "" {
			m := map[string]interface{}{
				"status":  400,
				"code":    "invalid_request",
				"message": "no Docker image specified",
			}

			b, err := json.Marshal(m)

			return string(b), err
		}

		b, err := json.Marshal(map[string]interface{}{
			"status":  200,
			"code":    "success",
			"message": "ECS Service successfully deployed",
		})

		return string(b), err
	})
}
