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
	// Pin an explicit version, not :latest. LocalStack's :latest (and :s3-latest) rolled
	// forward to Pro-gated builds that refuse to start without a LOCALSTACK_AUTH_TOKEN
	// (exit code 55 / "image no longer available"), which silently skips every S3 test.
	// 3.8 is the last widely-used community line that boots token-free.
	req := testcontainers.ContainerRequest{
		Image:        "localstack/localstack:3.8",
		ExposedPorts: []string{"4566/tcp"},
		NetworkMode:  testcontainers.Bridge,
		Env:          map[string]string{"SERVICES": "s3"},
		WaitingFor: wait.ForHTTP("/_localstack/health").WithPort("4566/tcp").WithStatusCodeMatcher(func(status int) bool {
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
