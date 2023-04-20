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

type ContainerImage struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Deployer interface {
	UpdateTask(ctx context.Context, family string, images []ContainerImage) error
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

func (d *deployer) UpdateTask(ctx context.Context, family string, images []ContainerImage) error {
	defOut, err := d.ecs.ListTaskDefinitions(ctx, &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: aws.String(family),
		MaxResults:   aws.Int32(1),
		Sort:         types.SortOrderDesc,
		Status:       types.TaskDefinitionStatusActive,
	})
	if err != nil {
		return err
	}
	if len(defOut.TaskDefinitionArns) == 0 {
		return errors.Wrap(TaskDefinitionNotFound, family)
	}

	out, err := d.ecs.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(defOut.TaskDefinitionArns[0]),
	})
	if err != nil {
		return err
	}

	// build new task definition
	reg := &ecs.RegisterTaskDefinitionInput{
		Family: aws.String(family),

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
	for _, c := range images {
		for i, d := range reg.ContainerDefinitions {
			if *d.Name == c.Name {
				reg.ContainerDefinitions[i].Image = aws.String(c.Image)
				changed++
				break
			}
		}
	}

	if len(images) != changed {
		return errors.Wrap(TaskDefinitionContainerMismatch, family)
	}

	// register new task definition
	_, err = d.ecs.RegisterTaskDefinition(ctx, reg)
	if err != nil {
		return err
	}

	// deregister old task definition
	_, err = d.ecs.DeregisterTaskDefinition(ctx, &ecs.DeregisterTaskDefinitionInput{
		TaskDefinition: out.TaskDefinition.TaskDefinitionArn,
	})
	if err != nil {
		return err
	}

	return nil
}
