package main

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	. "github.com/kai5263499/rhema"
	"github.com/kai5263499/rhema/domain"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := &domain.Config{}
	if err := env.Parse(cfg); err != nil {
		log.WithError(err).Fatal("parse config")
	}

	if level, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("parse log level")
	} else {
		log.SetLevel(level)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})

	logrus.SetReportCaller(true)

	speedupAudo := NewSpeedupAudio(cfg, exec.Command)
	scrape := NewScrape(cfg)
	text2mp3 := NewText2Mp3(cfg)
	youtube := NewYoutube(cfg, scrape, speedupAudo)

	contentStorage, err := NewContentStorage(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("new storage client")
	}

	requestProcessor := NewRequestProcessor(cfg,
		scrape,
		youtube,
		text2mp3,
		speedupAudo,
		contentStorage,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	api, err := NewApi(
		ctx,
		stop,
		cfg,
		requestProcessor,
		contentStorage,
	)
	if err != nil {
		logrus.WithError(err).Fatal("new api")
	}

	api.Start()

	<-ctx.Done()

	log.Info("api exiting")
}
