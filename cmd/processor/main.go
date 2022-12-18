package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"github.com/caarlos0/env/v6"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"

	. "github.com/kai5263499/rhema"
	"github.com/kai5263499/rhema/domain"
)

var (
	cfg              *domain.Config
	requestProcessor *RequestProcessor
)

func main() {
	cfg = &domain.Config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithError(err).Fatal("parse configs")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	logrus.SetReportCaller(true)

	speedupAudo := NewSpeedupAudio(cfg, exec.Command)
	scrape := NewScrape(cfg)
	text2mp3 := NewText2Mp3(cfg)
	youtube := NewYoutube(cfg, scrape, speedupAudo)

	redisConnStr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	logrus.Debugf("connecting to redis %s", redisConnStr)
	redisConn, redisConnErr := redis.Dial("tcp", redisConnStr)
	if redisConnErr != nil {
		logrus.WithError(redisConnErr).Fatal("unable to connect to redis")
	}

	requestProcessor = NewRequestProcessor(cfg, scrape, youtube, text2mp3, speedupAudo, redisConn)

	logrus.Info("finished setup, listening for messages from mqtt")

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
