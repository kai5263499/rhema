package rhema

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var _ domain.Comms = (*MqttComms)(nil)

func NewMqttComms(clientID string, mqttBroker string) (*MqttComms, error) {

	opts := mqtt.NewClientOptions().AddBroker(mqttBroker).SetClientID(clientID)
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal("mqtt new client")
		return nil, token.Error()
	}

	mc := &MqttComms{
		mqttClient:  mqttClient,
		requestChan: make(chan pb.Request, 100),
	}

	if token := mc.mqttClient.Subscribe(domain.RequestsTopic, 0, mc.messageHandler); token.Wait() && token.Error() != nil {
		logrus.WithError(token.Error()).Fatal("mqtt subscribe")
		return nil, token.Error()
	}

	return mc, nil
}

type MqttComms struct {
	mqttClient  mqtt.Client
	requestChan chan pb.Request
}

func (m *MqttComms) messageHandler(client mqtt.Client, msg mqtt.Message) {
	logrus.Debugf("got message with %d bytes from %s", len(msg.Payload()), msg.Topic())

	var req pb.Request
	if err := proto.Unmarshal(msg.Payload(), &req); err != nil {
		logrus.WithError(err).Errorf("unable to unmarshal request")
		return
	}

	m.requestChan <- req
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
	logrus.Debugf("published %d bytes to %s", len(pubBytes), domain.RequestsTopic)

	return nil
}

func (m *MqttComms) Close() {
	m.mqttClient.Disconnect(100)
}
