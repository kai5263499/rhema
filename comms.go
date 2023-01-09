package rhema

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kai5263499/rhema/domain"
	pb "github.com/kai5263499/rhema/generated"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

var _ domain.Comms = (*comms)(nil)

func NewComms(cfg *domain.Config) (retComms domain.Comms, err error) {

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
		"group.id":          cfg.KafkaGroupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return
	}

	consumer.SubscribeTopics([]string{cfg.KafkaRequestTopic}, nil)

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.KafkaBrokers,
	})
	if err != nil {
		return
	}

	retComms = &comms{
		cfg:           cfg,
		kafkaConsumer: consumer,
		kafkaProducer: producer,
		requestChan:   make(chan pb.Request, 100),
	}

	return
}

type comms struct {
	cfg           *domain.Config
	requestChan   chan pb.Request
	kafkaConsumer *kafka.Consumer
	kafkaProducer *kafka.Producer
}

func (m *comms) RequestChan() chan pb.Request {
	return m.requestChan
}

func (m *comms) SendRequest(req *pb.Request) (err error) {
	pubBytes, err := proto.Marshal(req)
	if err != nil {
		logrus.WithError(err).Errorf("unable to marshal proto %+#v", req)
		return
	}

	if err = m.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &m.cfg.KafkaRequestTopic, Partition: kafka.PartitionAny},
		Value:          pubBytes,
	}, nil); err != nil {
		return
	}

	logrus.Debugf("published %d bytes to %s", len(pubBytes), m.cfg.KafkaRequestTopic)

	return
}

func (m *comms) Close() (err error) {
	m.kafkaProducer.Close()

	if err = m.kafkaConsumer.Close(); err != nil {
		return
	}
	return
}

func (m *comms) consumeFromKafka() {
	for {
		msg, err := m.kafkaConsumer.ReadMessage(-1)
		if err == nil {
			logrus.Debugf("Message received: %s\n", string(msg.Value))

			if err := m.messageHandler(msg); err != nil {
				logrus.WithError(err).Errorf("error handling message")
				return
			}
		} else {
			logrus.WithError(err).Errorf("error recieving message")
			return
		}
	}
}

func (m *comms) messageHandler(msg *kafka.Message) (err error) {
	logrus.Debugf("got message with %d bytes from %s", len(msg.Value), *msg.TopicPartition.Topic)

	var req pb.Request

	if err = proto.Unmarshal(msg.Value, &req); err != nil {
		logrus.WithError(err).Errorf("unable to unmarshal request")
		return
	}

	m.requestChan <- req

	return
}
