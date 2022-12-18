package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/gomodule/redigo/redis"
	. "github.com/kai5263499/rhema"
	"github.com/kai5263499/rhema/domain"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

var (
	cfg *domain.Config
)

func main() {
	cfg = &domain.Config{}
	if err := env.Parse(&cfg); err != nil {
		log.WithError(err).Fatal("parse config")
	}

	if level, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("parse log level")
	} else {
		log.SetLevel(level)
	}

	logrus.SetReportCaller(true)

	redisConnStr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	log.Debugf("connecting to redis %s", redisConnStr)
	redisConn, redisConnErr := redis.Dial("tcp", redisConnStr)
	if redisConnErr != nil {
		log.WithError(redisConnErr).Fatal("unable to connect to redis")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	api := NewApi(
		ctx,
		stop,
		cfg,
		redisConn,
	)

	api.Start()

	<-ctx.Done()

	log.Info("api exiting")
}
