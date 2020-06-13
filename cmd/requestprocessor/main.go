package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	. "github.com/kai5263499/rhema"
)

type config struct {
	MQTTBroker            string  `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
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
	mqttClient       mqtt.Client
)

var mqttMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var req pb.Request
	if err := proto.Unmarshal(msg.Payload(), &req); err != nil {
		logrus.WithError(err).Errorf("request unmarshal error")
		return
	}

	if _, err := requestProcessor.Process(req); err != nil {
		logrus.WithError(err).Errorf("error processing request")
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

	opts := mqtt.NewClientOptions().AddBroker(cfg.MQTTBroker).SetClientID("requestprocessor")
	opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal("mqtt new client")
	}

	if token := mqttClient.Subscribe(domain.RequestsTopic, 0, nil); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatalf("error subscribing to topic %s", domain.RequestsTopic)
	}

	speedupAudo := NewSpeedupAudio(cfg.TmpPath, cfg.DefaultAtempo)
	scrape := NewScrape(uint32(cfg.MinTextBlockSize), cfg.TmpPath, cfg.TitleLengthLimit)
	text2mp3 := NewText2Mp3(cfg.TmpPath, cfg.DefaultWordsPerMinute, cfg.DefaultEspeakVoice)
	youtube := NewYoutube(scrape, speedupAudo, cfg.TmpPath)
	mqttComms := NewMqttComms(mqttClient)
	requestProcessor = NewRequestProcessor(cfg.TmpPath, scrape, youtube, text2mp3, speedupAudo, cfg.TitleLengthLimit, mqttComms)

	logrus.Infof("finished setup, listening for messages from mqtt")

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
