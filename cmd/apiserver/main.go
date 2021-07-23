package main

import (
	"fmt"
	"os"
	"os/signal"

	// External
	"github.com/caarlos0/env"
	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	// Internal
	. "github.com/kai5263499/rhema"
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

	api := NewApi(
		redisConn,
		&redisG,
		cfg.RedisGraphKey,
		mqttComms,
		cfg.Port,
		cfg.Auth0ClientId,
		cfg.Auth0ClientSecret,
		cfg.Auth0CallbackUrl,
		cfg.Auth0Domain,
		cfg.EndableGraphiql,
		cfg.SubmittedWith,
	)

	api.Setup()

	api.Start()

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
