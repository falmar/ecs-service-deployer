FROM golang:1.19-alpine as builder
WORKDIR /go-app
COPY go.mod go.sum ./
RUN go mod download
COPY /cmd /go-app/cmd
COPY /internal /go-app/internal
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o ./main ./cmd/lambda && \
    chmod +x ./main

FROM public.ecr.aws/lambda/go:1.2023.04.18.01
COPY --from=builder /go-app/main /var/task/main
CMD ["main"]
