# ECS Service Deployer

ECS Service Deployer is a Golang application that helps you deploy AWS ECS services. It is designed to be used as an AWS
Lambda function. While it is up to you how to use this Lambda function, a recommended use case is to invoke it from GitHub Actions CI/CD pipelines to automate your ECS service deployment process.

## Table of Contents
1. [Requirements](#requirements)
2. [Docker Images](#docker-images)
3. [Getting Started](#getting-started)
4. [Building the AWS Lambda Image](#building-the-aws-lambda-image)
5. [Local Testing](#local-testing)
6. [Contributing](#contributing)
7. [License](#license)
8. [Motivation](#motivation)
9. [TODO](#todo)

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

## Local Testing

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


## Contributing

If you have suggestions or improvements, feel free to create a pull request or open an issue on
the [GitHub repository](https://github.com/falmar/ecs-service-deployer).

## License

This project is licensed under the MIT License.

## Motivation

Why build my own ECS service deployer when there are already better tools and solutions out there? Well, the answer is simple: for the joy of learning by reinventing the wheel! This project was created to serve my own workflows. So, whether you're using this tool or just browsing the code, I hope you find it helpful or, at the very least, entertaining. Remember, there's never a wrong time to learn something new!

## TODO:

---

- [ ] Add which aws permissions are needed
