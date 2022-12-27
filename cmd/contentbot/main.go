package main

import (
	"os"
	"os/signal"

	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"

	. "github.com/kai5263499/rhema"
	"github.com/kai5263499/rhema/domain"
)

var (
	cfg   *domain.Config
	bot   *Bot
	comms domain.Comms
)

func main() {
	var err error
	cfg = &domain.Config{}
	if err = env.Parse(cfg); err != nil {
		logrus.WithError(err).Fatal("parse configs")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	logrus.SetReportCaller(true)

	comms, err = NewComms(cfg)
	if err != nil {
		logrus.WithError(err).Fatal("error creating comms")
	}

	bot, err = NewBot(cfg, comms)
	if err != nil {
		logrus.WithError(err).Fatal("error creating bot")
	}
	bot.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	func() {
		for sig := range c {
			// sig is a ^C, handle it
			logrus.Infof("got signal %d, exiting", sig)
			os.Exit(0)
		}
	}()
}
