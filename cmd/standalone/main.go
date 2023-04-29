package main

import (
	"context"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/falmar/ecs-service-deployer/internal"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

func initConfig(l *zap.Logger) func() {
	return func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")

		if err := viper.ReadInConfig(); err != nil {
			l.Fatal("Error reading config file", zap.Error(err))
		}

		l.Info("Using config file", zap.String("path", viper.ConfigFileUsed()))
		viper.AutomaticEnv()
	}
}

func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("containers", "c", []string{}, "Container images to update eg: (-c con1=image1 -c con2=image2)")
	cmd.Flags().String("task", "", "ECS Task Definition family")
	cmd.Flags().String("service", "", "ECS Service name or ARN")
	cmd.Flags().String("cluster", "", "ECS Service's Cluster Name or ARN")
	cmd.Flags().String("region", "", "AWS Region")
	cmd.Flags().BoolP("deregister", "d", false, "Deregister old task definition")

	_ = viper.BindPFlag("containers", cmd.Flags().Lookup("containers"))
	_ = viper.BindPFlag("ecs.task", cmd.Flags().Lookup("task"))
	_ = viper.BindPFlag("ecs.service", cmd.Flags().Lookup("service"))
	_ = viper.BindPFlag("ecs.cluster", cmd.Flags().Lookup("cluster"))
	_ = viper.BindPFlag("aws.region", cmd.Flags().Lookup("region"))
	_ = viper.BindPFlag("ecs.deregister", cmd.Flags().Lookup("deregister"))
}

func main() {
	var lConfig zap.Config
	if os.Getenv("DEBUG") != "" {
		lConfig = zap.NewDevelopmentConfig()
		lConfig.DisableCaller = true
		lConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		lConfig = zap.NewProductionConfig()
	}

	logger, _ := lConfig.Build()

	cobra.OnInitialize(initConfig(logger))
	ctx := context.Background()

	var rootCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deployer is a lambda function that deploys a ECS Service",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// init flags
			var containers []internal.ContainerImage
			c := viper.GetStringSlice("containers")
			task := viper.GetString("ecs.task")
			svc := viper.GetString("ecs.service")
			cluster := viper.GetString("ecs.cluster")
			deregister := viper.GetBool("ecs.deregister")

			if region := viper.GetString("aws.region"); region == "" {
				return errors.New("no AWS Region specified")
			}
			if task == "" {
				return errors.New("no ECS Task Definition family specified")
			}
			if svc == "" {
				return errors.New("no ECS Service specified")
			}
			if cluster == "" {
				return errors.New("no ECS Cluster specified")
			}
			if len(c) == 0 {
				return errors.New("no containers specified")
			}
			for _, pair := range c {
				container := strings.Split(pair, "=")

				containers = append(containers, internal.ContainerImage{
					Name:  container[0],
					Image: container[1],
				})
			}

			// init AWS config
			var options []func(*awsconfig.LoadOptions) error
			if viper.GetString("aws.region") != "" {
				options = append(options, awsconfig.WithRegion(viper.GetString("aws.region")))
			}
			if viper.GetString("aws.access_key_id") != "" &&
				viper.GetString("aws.secret_access_key") != "" {
				options = append(options, awsconfig.WithCredentialsProvider(
					credentials.NewStaticCredentialsProvider(
						viper.GetString("aws.access_key_id"),
						viper.GetString("aws.secret_access_key"),
						"",
					),
				))
			}

			awsConfig, err := awsconfig.LoadDefaultConfig(ctx, options...)
			if err != nil {
				return errors.Wrap(err, "error loading AWS config")
			}

			dp := internal.NewDeployer(internal.DeployerConfig{
				ECSClient: ecs.NewFromConfig(awsConfig),
			})

			// update task
			taskDefinition, err := dp.UpdateTask(ctx, &internal.UpdateTaskInput{
				Family:     task,
				Images:     containers,
				Deregister: deregister,
			})
			if err != nil {
				return err
			}

			// update service
			service, err := dp.DeployService(ctx, &internal.DeployServiceInput{
				Cluster:        cluster,
				Service:        svc,
				TaskDefinition: taskDefinition,
			})
			if err != nil {
				return err
			}

			logger.Info(
				"Service deployed",
				zap.String("service", *service.ServiceName),
				zap.String("task", *taskDefinition.Family),
			)

			return nil
		},
	}
	setFlags(rootCmd)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.Fatal("Error executing command", zap.Error(err))
	}
	os.Exit(0)
}
