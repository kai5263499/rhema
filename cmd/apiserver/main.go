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
	rg "github.com/redislabs/redisgraph-go"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type config struct {
	LogLevel                        string `env:"LOG_LEVEL" envDefault:"info"`
	SubmittedWith                   string `env:"SUBMITTED_WITH" envDefault:"api"`
	Port                            int    `env:"PORT" envDefault:"8080"`
	MQTTBroker                      string `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	MQTTClientID                    string `env:"MQTT_CLIENT_ID" envDefault:"contentbot"`
	RedisHost                       string `env:"REDIS_HOST"`
	RedisPort                       string `env:"REDIS_PORT" envDefault:"6379"`
	RedisGraphKey                   string `env:"REDIS_GRAPH_KEY" envDefault:"rhema-content"`
	GoogleServiceAccountKeyFilePath string `env:"GOOGLE_APPLICATION_CREDENTIALS"`
	Auth0ClientId                   string `env:"AUTH0_CLIENT_ID"`
	Auth0ClientSecret               string `env:"AUTH0_CLIENT_SECRET"`
	Auth0Domain                     string `env:"AUTH0_DOMAIN"`
	Auth0CallbackUrl                string `env:"AUTH0_CALLBACK_URL"`
	EndableGraphiql                 bool   `env:"ENABLE_GRAPHIQL" envDefault:"false"`
}

var (
	cfg config
)

func main() {
	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		log.WithError(err).Fatal("parse config")
	}

	if level, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("parse log level")
	} else {
		log.SetLevel(level)
	}

	logrus.SetReportCaller(true)

	mqttComms, mqttCommsErr := NewMqttComms(cfg.MQTTClientID, cfg.MQTTBroker)
	if mqttCommsErr != nil {
		log.WithError(mqttCommsErr).Fatal("new mqtt comms")
	}

	redisConnStr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	log.Debugf("connecting to redis %s", redisConnStr)
	redisConn, redisConnErr := redis.Dial("tcp", redisConnStr)
	if redisConnErr != nil {
		log.WithError(redisConnErr).Fatal("unable to connect to redis")
	}

	redisG := rg.GraphNew(cfg.RedisGraphKey, redisConn)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	api := NewApi(
		ctx,
		stop,
		redisConn,
		&redisG,
		cfg.RedisGraphKey,
		mqttComms,
		cfg.Port,
		cfg.Auth0ClientId,
		cfg.Auth0ClientSecret,
		cfg.Auth0CallbackUrl,
		cfg.Auth0Domain,
		cfg.SubmittedWith,
	)

	api.Start()

	<-ctx.Done()

	log.Info("api exiting")
}
