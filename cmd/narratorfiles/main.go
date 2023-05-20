package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/projmayhem/narratorfiles/cmd/narratorfiles/config"
	"github.com/projmayhem/narratorfiles/cmd/narratorfiles/logger"
	"github.com/projmayhem/narratorfiles/cmd/narratorfiles/siphon"
	"github.com/projmayhem/narratorfiles/cmd/narratorfiles/webui"
	"gitlab.com/gopkgz/handlers"
)

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lgr := logger.New()
	rpttr := &logger.ErrorReporter{Logger: lgr}

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
	siphon := siphon.New(ctx, svc, bucket, prefix)

	mux := http.NewServeMux()
	mux.HandleFunc("/", siphon.ListObjects)
	mux.Handle("/object/", http.StripPrefix("/object/", http.HandlerFunc(siphon.GetObject)))
	mux.Handle("/play/", http.StripPrefix("/play/", http.HandlerFunc(siphon.PlayAudio)))
	mux.HandleFunc("/static/", webui.StaticFileHandleFunc)

	server := &http.Server{
		Addr:    ":8082",
		Handler: handlers.LogMiddleware(handlers.PanicRecoveryMiddleware(mux, rpttr.ReportError), lgr.Printf),
	}

	return server.ListenAndServe()
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
