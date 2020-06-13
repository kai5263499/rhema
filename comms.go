package rhema

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var _ domain.Comms = (*MqttComms)(nil)

func NewMqttComms(mqttClient mqtt.Client) *MqttComms {
	return &MqttComms{
		mqttClient:  mqttClient,
		requestChan: make(chan pb.Request, 100),
	}
}

type MqttComms struct {
	mqttClient  mqtt.Client
	requestChan chan pb.Request
}

func (m *MqttComms) RequestChan() chan pb.Request {
	return m.requestChan
}

func (m *MqttComms) SendRequest(req pb.Request) error {
	pubBytes, err := proto.Marshal(&req)
	if err != nil {
		logrus.WithError(err).Errorf("unable to marshal proto")
		return err
	}

	if token := m.mqttClient.Publish(domain.RequestsTopic, 0, false, pubBytes); token.Error() != nil {
		logrus.WithError(token.Error()).Errorf("error sending request to marshal proto")
		return token.Error()
	}

	return nil
}
