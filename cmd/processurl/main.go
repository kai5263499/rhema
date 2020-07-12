package main

import (
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/gofrs/uuid"
	. "github.com/kai5263499/rhema"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

type config struct {
	MQTTBroker   string `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	MQTTClientID string `env:"MQTT_CLIENT_ID" envDefault:"processurl"`
	LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
}

var (
	cfg config
)

func main() {
	cfg = config{}
	if err := env.Parse(&cfg); err != nil {
		logrus.WithError(err).Fatal("parse config")
	}

	if level, err := logrus.ParseLevel(cfg.LogLevel); err != nil {
		logrus.WithError(err).Fatal("parse log level")
	} else {
		logrus.SetLevel(level)
	}

	mqttComms, mqttCommsErr := NewMqttComms(cfg.MQTTClientID, cfg.MQTTBroker)
	if mqttCommsErr != nil {
		logrus.WithError(mqttCommsErr).Fatal("new mqtt comms")
	}

	for _, arg := range os.Args[1:] {
		newUUID := uuid.Must(uuid.NewV4())

		req := pb.Request{
			Title:       newUUID.String(),
			Type:        pb.ContentType_URI,
			Created:     uint64(time.Now().Unix()),
			Uri:         arg,
			RequestHash: newUUID.String(),
		}

		if err := mqttComms.SendRequest(req); err != nil {
			logrus.WithError(err).Error("process request error")
			continue
		}

		logrus.Infof("request successfully submitted to processor")
	}

	mqttComms.Close()
}
