package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/storage"
	"github.com/caarlos0/env"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	. "github.com/kai5263499/rhema"
)

type config struct {
	MQTTBroker                   string `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	Bucket                       string `env:"BUCKET"`
	TmpPath                      string `env:"TMP_PATH" envDefault:"/tmp"`
	ChownTo                      int    `env:"CHOWN_TO" envDefault:"1000"`
	LogLevel                     string `env:"LOG_LEVEL" envDefault:"info"`
	GoogleApplicationCredentials string `env:"GOOGLE_APPLICATION_CREDENTIALS"`
}

var (
	cfg            config
	contentStorage *ContentStorage
	mqttClient     mqtt.Client
)

var mqttMessageHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var req pb.Request
	if err := proto.Unmarshal(msg.Payload(), &req); err != nil {
		logrus.WithError(err).Errorf("request unmarshal error")
		return
	}

	if _, err := contentStorage.Store(req); err != nil {
		logrus.WithError(err).Errorf("error storing content")
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

	// mqtt.DEBUG = log.New(os.Stdout, "", 0)
	// mqtt.ERROR = log.New(os.Stdout, "", 0)
	opts := mqtt.NewClientOptions().AddBroker(cfg.MQTTBroker).SetClientID("requeststorage")
	opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := mqttClient.Subscribe(domain.RequestsTopic, 0, nil); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatalf("error subscribing to topic %s", domain.RequestsTopic)
	}

	ctx := context.Background()
	gcpClient, newGCPStorageErr := storage.NewClient(ctx)
	if newGCPStorageErr != nil {
		logrus.WithError(newGCPStorageErr).Fatal("new gcp storage client")
	}

	var newStorageErr error
	contentStorage, newStorageErr = NewContentStorage(cfg.TmpPath, cfg.Bucket, gcpClient)
	if newStorageErr != nil {
		logrus.WithError(newStorageErr).Fatal("new storage client")
	}

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
