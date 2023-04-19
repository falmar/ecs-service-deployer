package internal

import (
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

type Deployer interface {
}

type DeployerConfig struct {
}

func NewDeployer(config DeployerConfig) Deployer {
	return &deployer{}
}

type deployer struct {
}

func (d *deployer) describeTask() error {
	ecs.New(ecs.Options{}, nil)

	return nil
}
