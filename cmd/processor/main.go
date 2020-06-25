package main

import (
	"os"
	"os/signal"

	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"

	. "github.com/kai5263499/rhema"
)

type config struct {
	MQTTBroker            string  `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	MQTTClientID          string  `env:"MQTT_CLIENT_ID" envDefault:"requestprocessor"`
	MinTextBlockSize      int     `env:"MIN_TEXT_BLOCK_SIZE" envDefault:"100"`
	TmpPath               string  `env:"TMP_PATH" envDefault:"/tmp"`
	DefaultWordsPerMinute int     `env:"DEFAULT_WORDS_PER_MINUTE" envDefault:"350"`
	DefaultEspeakVoice    string  `env:"DEFAULT_ESPEAK_VOICE" envDefault:"f5"`
	DefaultAtempo         float32 `env:"DEFAULT_ATEMPO" envDefault:"2.0"`
	ChownTo               int     `env:"CHOWN_TO" envDefault:"1000"`
	LogLevel              string  `env:"LOG_LEVEL" envDefault:"info"`
	TitleLengthLimit      int     `env:"TITLE_LENGTH_LIMIT" envDefault:"40"`
}

var (
	cfg              config
	requestProcessor *RequestProcessor
	mqttComms        *MqttComms
)

func mqttReadLoop() {
	for req := range mqttComms.RequestChan() {
		requestProcessor.Process(req)
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

	speedupAudo := NewSpeedupAudio(cfg.TmpPath, cfg.DefaultAtempo)
	scrape := NewScrape(uint32(cfg.MinTextBlockSize), cfg.TmpPath, cfg.TitleLengthLimit)
	text2mp3 := NewText2Mp3(cfg.TmpPath, cfg.DefaultWordsPerMinute, cfg.DefaultEspeakVoice)
	youtube := NewYoutube(scrape, speedupAudo, cfg.TmpPath)

	var mqttCommsErr error
	mqttComms, mqttCommsErr = NewMqttComms(cfg.MQTTClientID, cfg.MQTTBroker)
	if mqttCommsErr != nil {
		logrus.WithError(mqttCommsErr).Fatal("new mqtt comms")
	}

	requestProcessor = NewRequestProcessor(cfg.TmpPath, scrape, youtube, text2mp3, speedupAudo, cfg.TitleLengthLimit, mqttComms)

	go mqttReadLoop()

	logrus.Infof("finished setup, listening for messages from mqtt")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	func() {
		for sig := range c {
			// sig is a ^C, handle it
			logrus.Infof("got signal %d, exiting", sig)

			mqttComms.Close()

			os.Exit(0)
		}
	}()
}
