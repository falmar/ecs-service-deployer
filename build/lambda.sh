#!/bin/bash

sudo docker build -f ./build/lambda.Dockerfile -t docker.io/falmar/ecs-deployer:latest .
