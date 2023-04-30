resource "aws_ecs_cluster" "ecs_deployer_test" {
  name = "ecs_deployer_test"

  tags = {
    Application = "ecs_deployer_test"
    Name = "ecs_deployer_test"
  }
}

resource "aws_ecs_service" "ecs_deployer_test" {
  name = "ecs_deployer_test"
  cluster         = aws_ecs_cluster.ecs_deployer_test.id
  task_definition = aws_ecs_task_definition.ecs_deployer_test.arn
  desired_count   = var.service_count

  tags = {
    Application = "ecs_deployer_test"
    Name = "ecs_deployer_test"
  }
}

resource "aws_ecs_task_definition" "ecs_deployer_test" {
  container_definitions = jsonencode([
    {
      name         = "test"
      image        = "docker.io/falmar/hostname"
      essential    = true
      cpu          = 256
      memory       = 512
      portMappings = [
        {
          containerPort = 3000
          hostPort      = 3000
          protocol      = "tcp"
        }
      ]
    }
  ])

  family                   = "ecs_deployer_test"
  cpu                      = "256"
  memory                   = "512"

  tags = {
    Application = "ecs_deployer_test"
    Name = "ecs_deployer_test"
  }
}
