package internal

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/pkg/errors"
)

var TaskDefinitionNotFound = errors.New("no family definitions found")
var TaskDefinitionContainerMismatch = errors.New("number of containers in task definition does not match number of containers in update")
var ServiceNotFound = errors.New("service not found")

type UpdateTaskInput struct {
	Family string
	Images []ContainerImage
}

type DeployServiceInput struct {
	Cluster        string
	Service        string
	TaskDefinition *types.TaskDefinition
}

type ContainerImage struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Deployer interface {
	UpdateTask(ctx context.Context, input *UpdateTaskInput) (*types.TaskDefinition, error)
	DeployService(ctx context.Context, input *DeployServiceInput) (*types.Service, error)
}

type DeployerConfig struct {
	ECSClient *ecs.Client
}

func NewDeployer(config DeployerConfig) Deployer {
	return &deployer{
		ecs: config.ECSClient,
	}
}

type deployer struct {
	ecs *ecs.Client
}

func (d *deployer) UpdateTask(ctx context.Context, input *UpdateTaskInput) (*types.TaskDefinition, error) {
	defOut, err := d.ecs.ListTaskDefinitions(ctx, &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: aws.String(input.Family),
		MaxResults:   aws.Int32(1),
		Sort:         types.SortOrderDesc,
		Status:       types.TaskDefinitionStatusActive,
	})
	if err != nil {
		return nil, err
	}
	if len(defOut.TaskDefinitionArns) == 0 {
		return nil, errors.Wrap(TaskDefinitionNotFound, input.Family)
	}

	out, err := d.ecs.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(defOut.TaskDefinitionArns[0]),
	})
	if err != nil {
		return nil, err
	}

	// build new task definition
	reg := &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(input.Family),

		ContainerDefinitions:    out.TaskDefinition.ContainerDefinitions,
		Cpu:                     out.TaskDefinition.Cpu,
		EphemeralStorage:        out.TaskDefinition.EphemeralStorage,
		ExecutionRoleArn:        out.TaskDefinition.ExecutionRoleArn,
		InferenceAccelerators:   out.TaskDefinition.InferenceAccelerators,
		IpcMode:                 out.TaskDefinition.IpcMode,
		Memory:                  out.TaskDefinition.Memory,
		NetworkMode:             out.TaskDefinition.NetworkMode,
		PidMode:                 out.TaskDefinition.PidMode,
		PlacementConstraints:    out.TaskDefinition.PlacementConstraints,
		ProxyConfiguration:      out.TaskDefinition.ProxyConfiguration,
		RequiresCompatibilities: out.TaskDefinition.RequiresCompatibilities,
		RuntimePlatform:         out.TaskDefinition.RuntimePlatform,
		TaskRoleArn:             out.TaskDefinition.TaskRoleArn,
		Volumes:                 out.TaskDefinition.Volumes,
	}
	// can only update tags if they exist
	if out.Tags != nil && len(out.Tags) > 0 {
		reg.Tags = out.Tags
	}

	// update container images
	var changed int
	for _, c := range input.Images {
		for i, d := range reg.ContainerDefinitions {
			if *d.Name == c.Name {
				reg.ContainerDefinitions[i].Image = aws.String(c.Image)
				changed++
				break
			}
		}
	}

	if len(input.Images) != changed {
		return nil, errors.Wrap(TaskDefinitionContainerMismatch, input.Family)
	}

	// register new task definition
	td, err := d.ecs.RegisterTaskDefinition(ctx, reg)
	if err != nil {
		return nil, err
	}

	// deregister old task definition
	_, err = d.ecs.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: out.TaskDefinition.TaskDefinitionArn,
	})
	if err != nil {
		return nil, err
	}

	return td.TaskDefinition, nil
}

func (d *deployer) DeployService(ctx context.Context, input *DeployServiceInput) (*types.Service, error) {
	outDS, err := d.ecs.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  aws.String(input.Cluster),
		Services: []string{input.Service},
	})
	if err != nil {
		return nil, err
	}
	if len(outDS.Services) == 0 {
		return nil, ServiceNotFound
	}

	// get current service connect config if any
	var connectConfig *types.ServiceConnectConfiguration = nil
	if len(outDS.Services[0].Deployments) > 0 {
		for _, deployment := range outDS.Services[0].Deployments {
			if deployment.TaskDefinition == input.TaskDefinition.TaskDefinitionArn {
				connectConfig = deployment.ServiceConnectConfiguration
				break
			}
		}
	}

	// update service
	outUS, err := d.ecs.UpdateService(ctx, &ecs.UpdateServiceInput{
		Cluster:        outDS.Services[0].ClusterArn,
		Service:        outDS.Services[0].ServiceArn,
		TaskDefinition: input.TaskDefinition.TaskDefinitionArn,

		CapacityProviderStrategy:      outDS.Services[0].CapacityProviderStrategy,
		DeploymentConfiguration:       outDS.Services[0].DeploymentConfiguration,
		DesiredCount:                  aws.Int32(outDS.Services[0].DesiredCount),
		EnableECSManagedTags:          aws.Bool(outDS.Services[0].EnableECSManagedTags),
		EnableExecuteCommand:          aws.Bool(outDS.Services[0].EnableExecuteCommand),
		ForceNewDeployment:            true,
		HealthCheckGracePeriodSeconds: outDS.Services[0].HealthCheckGracePeriodSeconds,
		LoadBalancers:                 outDS.Services[0].LoadBalancers,
		NetworkConfiguration:          outDS.Services[0].NetworkConfiguration,
		PlacementConstraints:          outDS.Services[0].PlacementConstraints,
		PlacementStrategy:             outDS.Services[0].PlacementStrategy,
		PlatformVersion:               outDS.Services[0].PlatformVersion,
		PropagateTags:                 outDS.Services[0].PropagateTags,
		ServiceRegistries:             outDS.Services[0].ServiceRegistries,

		ServiceConnectConfiguration: connectConfig,
	})
	if err != nil {
		return nil, err
	}

	return outUS.Service, nil
}
