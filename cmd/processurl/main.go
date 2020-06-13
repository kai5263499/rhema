package main

import (
	"os"
	"time"

	"github.com/caarlos0/env"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofrs/uuid"
	. "github.com/kai5263499/rhema"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
)

type config struct {
	MQTTBroker string `env:"MQTT_BROKER" envDefault:"tcp://172.17.0.3:1883"`
	LogLevel   string `env:"LOG_LEVEL" envDefault:"info"`
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

	opts := mqtt.NewClientOptions().AddBroker(cfg.MQTTBroker).SetClientID("processurl")
	// opts.SetDefaultPublishHandler(mqttMessageHandler)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal("mqtt new client")
	}

	mqttComms := NewMqttComms(mqttClient)

	for _, arg := range os.Args[1:] {
		newUUID := uuid.Must(uuid.NewV4())

		req := pb.Request{
			Title:       newUUID.String(),
			Type:        pb.Request_URI,
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
}
