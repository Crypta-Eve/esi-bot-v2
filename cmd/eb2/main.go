package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eveisesi/eb2"
	"github.com/eveisesi/eb2/server"
	"github.com/eveisesi/eb2/slack"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

var (
	logger = logrus.New()
	cfg    eb2.Config
	err    error
)

func main() {
	err = getEnvConfig()
	if err != nil {
		fmt.Printf("failed to read env variables: %s", err)
	}

	logger.SetOutput(os.Stdout)

	logger.SetFormatter(&logrus.JSONFormatter{})

	loglvl, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.WithError(err).Fatal("failed to read in log level")
	}

	logger.SetLevel(loglvl)

	slackServ := slack.New(logger, &cfg)

	server := server.NewServer(&cfg, logger, slackServ)

	errChan := make(chan error, 1)

	signals := make(chan os.Signal, 1)

	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		errChan <- server.Run()
	}()

	select {
	case err := <-errChan:
		logger.WithError(err).Fatal("error encountered attempting to start the http server")
	case <-signals:
		logger.Info("starting server shutdown procedure....")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = server.Shutdown(ctx)
		if err != nil {
			logger.WithError(err).Panic("unable to shutdown the http server")
		}
	}

}

func getEnvConfig() error {
	return envconfig.Process("", &cfg)
}
