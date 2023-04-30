package standalone

import (
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/falmar/ecs-service-deployer/internal"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"strings"
)

func initConfig(l *zap.Logger) func() {
	return func() {
		configPath := viper.GetString("config")
		if configPath != "" {
			viper.SetConfigFile(configPath)
		} else {
			viper.AddConfigPath("./config")
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
		}

		viper.AutomaticEnv()

		err := viper.ReadInConfig()
		if errors.Is(err, os.ErrNotExist) {
			if configPath != "" {
				l.Fatal("config file not found", zap.String("path", configPath))
			}

			l.Warn("config file not found, skipping")
			return
		} else if err != nil {
			l.Fatal("error reading config file", zap.Error(err))
		}

		l.Info("using config file", zap.String("path", viper.ConfigFileUsed()))
	}
}

func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("config", "c", "", "config file")
	cmd.Flags().StringArray("containers", []string{}, "Container images to update eg: (--containers con1=image1 --containers con2=image2)")
	cmd.Flags().String("task", "", "ECS Task Definition family")
	cmd.Flags().String("service", "", "ECS Service name or ARN")
	cmd.Flags().String("cluster", "", "ECS Service's Cluster Name or ARN")
	cmd.Flags().String("region", "", "AWS Region")
	cmd.Flags().BoolP("deregister", "d", false, "Deregister old task definition")

	_ = viper.BindPFlag("config", cmd.Flags().Lookup("config"))
	_ = viper.BindPFlag("containers", cmd.Flags().Lookup("containers"))
	_ = viper.BindPFlag("ecs.task", cmd.Flags().Lookup("task"))
	_ = viper.BindPFlag("ecs.service", cmd.Flags().Lookup("service"))
	_ = viper.BindPFlag("ecs.cluster", cmd.Flags().Lookup("cluster"))
	_ = viper.BindPFlag("aws.region", cmd.Flags().Lookup("region"))
	_ = viper.BindPFlag("ecs.deregister", cmd.Flags().Lookup("deregister"))
}

func Cmd(logger *zap.Logger) *cobra.Command {

	cobra.OnInitialize(initConfig(logger))

	var rootCmd = &cobra.Command{
		Use:           "deploy",
		Short:         "Deployer is a lambda function that deploys a ECS Service",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// init flags
			var containers []internal.ContainerImage
			c := viper.GetStringSlice("containers")
			task := viper.GetString("ecs.task")
			svc := viper.GetString("ecs.service")
			cluster := viper.GetString("ecs.cluster")
			deregister := viper.GetBool("ecs.deregister")

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

	return rootCmd
}
