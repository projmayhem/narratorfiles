package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

func ConfigFromEnv() (*aws.Config, error) {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	endpoint := os.Getenv("AWS_S3_ENDPOINT")
	region := os.Getenv("AWS_REGION")

	if accessKey == "" || secretKey == "" || endpoint == "" || region == "" {
		return nil, fmt.Errorf("missing required environment variables: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_S3_ENDPOINT, or AWS_REGION")
	}

	config := &aws.Config{
		Region:           aws.String(region),
		Endpoint:         aws.String(endpoint),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		S3ForcePathStyle: aws.Bool(true),
	}

	return config, nil
}

func BucketFromEnv() (string, error) {
	bucket := os.Getenv("AWS_S3_BUCKET")

	if bucket == "" {
		return "", errors.New("missing required environment variable: AWS_S3_BUCKET")
	}

	return bucket, nil
}

func PrefixFromEnv() string {
	return os.Getenv("AWS_S3_PREFIX")
}
