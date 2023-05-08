# ECS Service Deployer

ECS Service Deployer is a Golang application that helps you deploy AWS ECS services. It is designed to be used as an AWS
Lambda function. While it is up to you how to use this Lambda function, a recommended use case is to invoke it from GitHub Actions CI/CD pipelines to automate your ECS service deployment process.

## Table of Contents
1. [Requirements](#requirements)
2. [Docker Images](#docker-images)
3. [Getting Started](#getting-started)
4. [Building the AWS Lambda Image](#building-the-aws-lambda-image)
5. [Testing Lambda Locally](#testing-lambda-locally)
6. [Deployer CLI](#ecs-service-deployer-cli)
7. [AWS Permissions](#aws-permissions)
8. [Contributing](#contributing)
9. [License](#license)
10. [Motivation](#motivation)
11. [TODO](#todo)

## Requirements

To use this Golang application for deploying AWS ECS services, you will need the following:

- Go 1.19+
- Docker (to build the ECR image)
- AWS CLI (to log in to Docker and push the image)
- AWS Account
- AWS ECR registry

## Docker Images

Pre-built Docker images for the ECS Service Deployer are available on [Docker Hub](https://hub.docker.com/r/falmar/ecs-service-deployer). Images are provided for both Linux `amd64` and `arm64` architectures.

## Getting Started

To get started, clone the repository:

```bash
git clone https://github.com/falmar/ecs-service-deployer.git
cd ecs-service-deployer
```

## Building the AWS Lambda Image

You can build the AWS Lambda Docker image using the following command:

```bash
docker build -f ./build/lambda.Dockerfile -t <your-aws-ecr-registry-url>/ecs-service-deployer .
```

Replace <your-aws-ecr-registry-url> with your own AWS ECR registry URL. This command will use the lambda.Dockerfile to build the Docker image.

To create a image for a different platform, follow the instructions in the [Docker documentation](https://docs.docker.com/build/building/multi-platform/).

## Testing Lambda Locally

You can test the application locally using Docker. The `local.Dockerfile` is provided for this purpose.

1. Build the local Docker image:

```bash
docker build -f ./build/local.Dockerfile -t <your-aws-ecr-registry-url>/ecs-service-deployer:local .
```

Replace `<your-aws-ecr-registry-url>` with your own AWS ECR registry URL. This command will use the `local.Dockerfile` to build the local Docker image.

2. Run the local Docker container:

```bash
docker run --rm -it -p 9000:8080 -e DEBUG=1 -e AWS_LAMBDA_FUNCTION_MEMORY_SIZE=512 \
  <your-aws-ecr-registry-url>/ecs-service-deployer:local
```

The application will be accessible at `http://localhost:9000`. Make sure to set any required environment variables and configure your local AWS credentials for testing.

## ECS Service Deployer CLI

In addition to the Lambda function, you can also use the provided CLI to deploy your ECS services on demand. The CLI offers the same functionality as the Lambda function and can be executed locally.

### Usage

To use the ECS Service Deployer CLI, run the following command:

```bash
$ go run ./cmd/standalone --service=<ECS_SERVICE> --cluster=<ECS_CLUSTER> --task=<ECS_TASK_FAMILY> --containers <CONTAINER_NAME>=<CONTAINER_IMAGE>
```

Replace `<ECS_SERVICE>`, `<ECS_CLUSTER>`, `<ECS_TASK_FAMILY>`, `<CONTAINER_NAME>`, and `<CONTAINER_IMAGE>` with the appropriate values for your use case.


> NOTE: You can also install the CLI on your system by running the following command:

```bash
$ go install .
```

### Example

Here's an example of how to use the CLI to deploy an ECS service:

```bash
$ go run ./cmd/standalone --service=ecs_deployer_test --cluster=ecs_deployer_test --task=ecs_deployer_test --containers test=nginx:alpine
```

This command will deploy the `ecs_deployer_test` service on the `ecs_deployer_test` cluster, using the `ecs_deployer_test` task family and updating the container named `test` with the `nginx:alpine` image.

Make sure your AWS credentials and configuration are properly set up in your environment before using the CLI.

### Terraform template

A Terraform template is provided in the [./terraform/ecs.tf](./terraform/ecs.tf) file, which you can use to set up the required AWS resources for your ECS service. Make sure to customize the template according to your needs before using it.

```bash
$ terraform init
# change the service_count variable to the number of services you want to deploy or leave it as is to just create the resources
$ terraform apply --var service_count=0
```

## AWS Permissions

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "SvcWrite",
      "Effect": "Allow",
      "Action": [
        "ecs:UpdateService",
        "ecs:DescribeServices"
      ],
      "Resource": "arn:aws:ecs:<region>:<account>:service/<cluster>/<svc>",
      "Condition": {
        "ArnEquals": {
          "ecs:cluster": "arn:aws:ecs:<region>:<account>:cluster/<cluster>"
        },
        "StringEqualsIfExists": {
          "aws:ResourceTag/ECS_Deployer": "true"
        }
      }
    },
    {
      "Sid": "TaskRead",
      "Effect": "Allow",
      "Action": [
        "ecs:ListTaskDefinitions",
        "ecs:DescribeTaskDefinition",
        "ecs:DeregisterTaskDefinition"
      ],
      "Resource": "*"
    },
    {
      "Sid": "TaskWrite",
      "Effect": "Allow",
      "Action": [
        "ecs:RegisterTaskDefinition"
      ],
      "Resource": "*",
      "Condition": {
        "StringEqualsIfExists": {
          "aws:ResourceTag/ECS_Deployer": "true"
        }
      }
    }
  ]
}
```

Replace `<region>`, `<account>`, `<cluster>`, `<svc>`, and `<TASK_FAMILY>` with the appropriate values for your use case.

> "List, Describe and Deregister" for TaskDefinition doesn't allow conditions or resource level permissions [see here](https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonelasticcontainerservice.html), so they are allowed for all resources. 

## Contributing

If you have suggestions or improvements, feel free to create a pull request or open an issue on
the [GitHub repository](https://github.com/falmar/ecs-service-deployer).

## License

This project is licensed under the MIT License.

## Motivation

Why build my own ECS service deployer when there are already better tools and solutions out there? Well, the answer is simple: for the joy of learning by reinventing the wheel! This project was created to serve my own workflows. So, whether you're using this tool or just browsing the code, I hope you find it helpful or, at the very least, entertaining. Remember, there's never a wrong time to learn something new!

## TODO:

- [ ] Add which aws permissions are needed
