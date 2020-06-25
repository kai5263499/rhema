package main

import (
	"context"
	"os"
	"os/signal"

	"cloud.google.com/go/storage"
	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"

	. "github.com/kai5263499/rhema"
)

type config struct {
	MQTTBroker                   string `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	MQTTClientID                 string `env:"MQTT_CLIENT_ID" envDefault:"storage"`
	TmpPath                      string `env:"TMP_PATH" envDefault:"/tmp"`
	ChownTo                      int    `env:"CHOWN_TO" envDefault:"1000"`
	LogLevel                     string `env:"LOG_LEVEL" envDefault:"info"`
	CopyToCloud                  bool   `env:"COPY_TO_CLOUD" envDefault:"true"`
	Bucket                       string `env:"BUCKET"`
	GoogleApplicationCredentials string `env:"GOOGLE_APPLICATION_CREDENTIALS"`
	CopyTmpToLocal               bool   `env:"COPY_TMP_TO_LOCAL" envDefault:"true"`
	LocalPath                    string `env:"LOCAL_PATH" envDefault:"/data"`
}

var (
	cfg            config
	contentStorage *ContentStorage
	mqttComms      *MqttComms
)

func mqttReadLoop() {
	for req := range mqttComms.RequestChan() {
		contentStorage.Store(req)
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

	var mqttCommsErr error
	mqttComms, mqttCommsErr = NewMqttComms(cfg.MQTTClientID, cfg.MQTTBroker)
	if mqttCommsErr != nil {
		logrus.WithError(mqttCommsErr).Fatal("new mqtt comms")
	}

	ctx := context.Background()
	gcpClient, newGCPStorageErr := storage.NewClient(ctx)
	if newGCPStorageErr != nil {
		logrus.WithError(newGCPStorageErr).Fatal("new gcp storage client")
	}

	var newStorageErr error
	contentStorage, newStorageErr = NewContentStorage(cfg.TmpPath, cfg.Bucket, gcpClient, cfg.CopyTmpToLocal, cfg.LocalPath, cfg.ChownTo, cfg.CopyToCloud)
	if newStorageErr != nil {
		logrus.WithError(newStorageErr).Fatal("new storage client")
	}

	go mqttReadLoop()

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
