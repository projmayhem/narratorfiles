package main

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/projmayhem/narratorfiles/cmd/narratorfiles/config"
	"github.com/projmayhem/narratorfiles/cmd/narratorfiles/siphon"
)

func run() error {
	bucket, err := config.BucketFromEnv()
	if err != nil {
		return fmt.Errorf("failed to get bucket from env: %w", err)
	}
	prefix := config.PrefixFromEnv()

	awsCfg, err := config.ConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to get aws config from env: %w", err)
	}
	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return fmt.Errorf("failed to create aws session: %w", err)
	}
	svc := s3.New(sess)

	siphon := &siphon.Siphon{
		Client: svc,
		Bucket: bucket,
		Prefix: prefix,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", siphon.ListObjects)

	server := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}

	return server.ListenAndServe()
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
