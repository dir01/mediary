package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"

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
	_ = os.Setenv("AWS_ENDPOINT_URL_S3", endpoint)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
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
