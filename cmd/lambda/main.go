package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/falmar/ecs-service-deployer/internal"
	"go.uber.org/zap"
	"os"
)

type Event struct {
	Containers     []internal.ContainerImage `json:"containers"`
	Service        string                    `json:"service"`
	TaskDefinition string                    `json:"task_definition"`
	Cluster        string                    `json:"cluster"`
	Deregister     bool                      `json:"deregister,omitempty"`
}

func encodeMessage(d interface{}) (string, error) {
	b, err := json.Marshal(d)

	return string(b), err
}

func main() {
	var lConfig zap.Config
	if os.Getenv("DEBUG") != "" {
		lConfig = zap.NewDevelopmentConfig()
	} else {
		lConfig = zap.NewProductionConfig()
	}

	ctx := context.Background()
	logger, _ := lConfig.Build()

	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Fatal("Error loading AWS lConfig", zap.Error(err))
	}

	// credentials should be set by the Lambda runtime given the
	// IAM role, it should enough to narrow down the permissions to
	// the ECS cluster/service/task access
	ecsClient := ecs.NewFromConfig(awsConfig)
	dp := internal.NewDeployer(internal.DeployerConfig{
		ECSClient: ecsClient,
	})

	lambda.Start(func(ctx context.Context, e Event) (string, error) {
		// TODO: docker repository and image tag validation?
		if len(e.Containers) == 0 {
			b, err := encodeMessage(map[string]interface{}{
				"status":  400,
				"code":    "invalid_request",
				"message": "no Docker images specified",
			})
			logger.Warn("no Docker images specified")

			return b, err
		}
		if e.TaskDefinition == "" {
			b, err := encodeMessage(map[string]interface{}{
				"status":  400,
				"code":    "invalid_request",
				"message": "no ECS Task Definition specified",
			})
			logger.Warn("no ECS Task Definition specified")

			return b, err
		}
		if e.Service == "" {
			b, err := encodeMessage(map[string]interface{}{
				"status":  400,
				"code":    "invalid_request",
				"message": "no ECS Service specified",
			})
			logger.Warn("no ECS Service specified")

			return b, err
		}
		if e.Cluster == "" {
			b, err := encodeMessage(map[string]interface{}{
				"status":  400,
				"code":    "invalid_request",
				"message": "no ECS Cluster specified",
			})
			logger.Warn("no ECS Cluster specified")

			return b, err
		}

		logger := logger.With(
			zap.String("service", e.Service),
			zap.String("task", e.TaskDefinition),
			zap.String("cluster", e.Cluster),
		)

		// update task
		taskDefinition, err := dp.UpdateTask(ctx, &internal.UpdateTaskInput{
			Family:     e.TaskDefinition,
			Images:     e.Containers,
			Deregister: e.Deregister,
		})
		if err != nil {
			b, err := encodeMessage(map[string]interface{}{
				"status":  500,
				"code":    "internal_server_error",
				"message": fmt.Sprintf("error updating task definition: %s", err.Error()),
			})
			logger.Error("error updating task definition", zap.Error(err))

			return b, err
		}

		// update service
		_, err = dp.DeployService(ctx, &internal.DeployServiceInput{
			Cluster:        e.Cluster,
			Service:        e.Service,
			TaskDefinition: taskDefinition,
		})
		if err != nil {
			b, err := encodeMessage(map[string]interface{}{
				"status":  500,
				"code":    "internal_server_error",
				"message": fmt.Sprintf("error updating service: %s", err.Error()),
			})
			logger.Error("error updating service", zap.Error(err))

			return b, err
		}

		b, err := json.Marshal(map[string]interface{}{
			"status":  200,
			"code":    "success",
			"message": "ECS Service successfully deployed",
		})
		logger.Info(
			"ECS Service successfully deployed",
			zap.String("task", fmt.Sprintf("%s:%d", *taskDefinition.Family, taskDefinition.Revision)),
		)

		return string(b), err
	})
}
