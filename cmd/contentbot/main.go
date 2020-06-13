package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"

	. "github.com/kai5263499/rhema"
)

type config struct {
	MQTTBroker string   `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	SlackToken string   `env:"SLACK_TOKEN"`
	Channels   []string `env:"CHANNELS" envDefault:"content"`
	LogLevel   string   `env:"LOG_LEVEL" envDefault:"info"`
	TmpPath    string   `env:"TMP_PATH" envDefault:"/tmp"`
	ChownTo    int      `env:"CHOWN_TO" envDefault:"1000"`
}

var (
	cfg config
)

func main() {
	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithError(err).Fatal("parse configs")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	opts := mqtt.NewClientOptions().AddBroker(cfg.MQTTBroker).SetClientID("contentbot")
	// opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal("mqtt new client")
	}

	mqttComms := NewMqttComms(mqttClient)

	bot := NewBot(cfg.SlackToken, cfg.Channels, cfg.TmpPath, cfg.ChownTo, mqttComms)
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
