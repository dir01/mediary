package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func getS3Client(ctx context.Context, bucketName string) (*s3.S3, error) {
	var targetPort nat.Port = "9090"

	req := testcontainers.ContainerRequest{
		Image:        "adobe/s3mock",
		ExposedPorts: []string{fmt.Sprintf("%s/tcp", targetPort)},
		Env:          map[string]string{"debug": "true", "trace": "true", "initialBuckets": bucketName, "root": "/tmp/s3mock"},
		WaitingFor:   wait.ForListeningPort(targetPort).WithStartupTimeout(5 * time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, targetPort)
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:   aws.String("eu-central-1"),
		Endpoint: aws.String(fmt.Sprintf("http://%s:%s", ip, port.Port())),
	})
	s3Client := s3.New(sess, aws.NewConfig().WithS3ForcePathStyle(true))

	return s3Client, nil
}
