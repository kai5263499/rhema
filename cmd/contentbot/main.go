package main

import (
	"os"
	"os/signal"

	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"

	. "github.com/kai5263499/rhema"
)

type config struct {
	MQTTBroker     string   `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	MQTTClientID   string   `env:"MQTT_CLIENT_ID" envDefault:"contentbot"`
	SlackToken     string   `env:"SLACK_TOKEN"`
	SubmittedWith  string   `env:"SUBMITTED_WITH" envDefault:"contentbot"`
	Channels       []string `env:"CHANNELS" envDefault:"content"`
	LogLevel       string   `env:"LOG_LEVEL" envDefault:"info"`
	TmpPath        string   `env:"TMP_PATH" envDefault:"/tmp"`
	ChownTo        int      `env:"CHOWN_TO" envDefault:"1000"`
	ATempo         string   `env:"ATEMPO" envDefault:"2.0"`
	WordsPerMinute int      `env:"WORDS_PER_MINUTE" envDefault:"350"`
	ESpeakVoice    string   `env:"ESPEAK_VOICE" envDefault:"f5"`
}

var (
	cfg       config
	bot       *Bot
	mqttComms *MqttComms
)

func mqttReadLoop() {
	for req := range mqttComms.RequestChan() {
		bot.Process(req)
	}
}

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

	logrus.SetReportCaller(true)

	var newMqttCommsErr error
	mqttComms, newMqttCommsErr = NewMqttComms(cfg.MQTTClientID, cfg.MQTTBroker)
	if newMqttCommsErr != nil {
		logrus.WithError(newMqttCommsErr).Fatal("unable to create mqtt comms")
	}

	go mqttReadLoop()

	bot = NewBot(cfg.SlackToken, cfg.Channels, cfg.TmpPath, cfg.ChownTo, mqttComms, cfg.ATempo, cfg.WordsPerMinute, cfg.ESpeakVoice, cfg.SubmittedWith)
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
