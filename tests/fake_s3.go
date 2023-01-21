package tests

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func GetS3Client(ctx context.Context, bucketName string) (client *s3.Client, teardown func(), err error) {
	req := testcontainers.ContainerRequest{
		Image:        "localstack/localstack:latest",
		ExposedPorts: []string{"4566/tcp"},
		NetworkMode:  testcontainers.Bridge,
		WaitingFor: wait.ForHTTP("/").WithPort("4566/tcp").WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusOK
		}),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, func() {}, fmt.Errorf("error creating container: %w", err)
	}
	teardown = func() {
		if err := container.Terminate(ctx); err != nil {
			fmt.Printf("error terminating container: %v", err)
		}
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, teardown, err
	}

	port, err := container.MappedPort(ctx, "4566/tcp")
	if err != nil {
		return nil, teardown, err
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{URL: endpoint}, nil
		})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local environment",
			},
		}),
	)

	if err != nil {
		return nil, teardown, err
	}
	client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	if _, err = client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucketName)}); err != nil {
		return nil, teardown, err
	}

	return client, teardown, nil
}

//func GetS3Client(ctx context.Context, bucketName string) (*s3.S3, error) {
//	var targetPort nat.Port = "9090"
//
//	req := testcontainers.ContainerRequest{
//		Image:        "adobe/s3mock",
//		ExposedPorts: []string{fmt.Sprintf("%s/tcp", targetPort)},
//		Env:          map[string]string{"debug": "true", "trace": "true", "initialBuckets": bucketName, "root": "/tmp/s3mock"},
//		WaitingFor:   wait.ForListeningPort(targetPort).WithStartupTimeout(5 * time.Minute),
//	}
//
//	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
//		ContainerRequest: req,
//		Started:          true,
//	})
//	if err != nil {
//		return nil, fmt.Errorf("failed to start container: %w", err)
//	}
//
//	ip, err := container.Host(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get container host: %w", err)
//	}
//
//	port, err := container.MappedPort(ctx, targetPort)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get container port: %w", err)
//	}
//
//	sess, err := session.NewSession(&aws.Config{
//		Region:   aws.String("eu-central-1"),
//		Endpoint: aws.String(fmt.Sprintf("http://%s:%s", ip, port.Port())),
//	})
//	s3Client := s3.New(sess, aws.NewConfig().WithS3ForcePathStyle(true))
//
//	return s3Client, nil
//}
