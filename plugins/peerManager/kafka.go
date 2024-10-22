package peermanager

import (
	"fmt"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/openrelayxyz/cardinal-streams/transports"
)

func strPtr(x string) *string {
	return &x
}

func createProducer(broker, topic string) (sarama.AsyncProducer, error) {

	sessionBrokers, sessionKafkaConfig = transports.ParseKafkaURL(strings.TrimPrefix(broker, "kafka://"))
	configEntries := make(map[string]*string)
	configEntries["retention.ms"] = strPtr("3600000")

	if err := transports.CreateTopicIfDoesNotExist(strings.TrimPrefix(broker, "kafka://"), topic, 1, configEntries); err != nil {
		panic(fmt.Sprintf("Could not create topic %v on broker %v: %v", topic, broker, err.Error()))
	}

	producer, err := sarama.NewAsyncProducer(sessionBrokers, sessionKafkaConfig)
	if err != nil {
		panic(fmt.Sprintf("Could not setup producer, peer manager plugin: %v", err.Error()))
	}

	return producer, nil
}

func createConsumer(broker, topic string) (sarama.PartitionConsumer, error) {

	consumer, err := sarama.NewConsumer(sessionBrokers, sessionKafkaConfig)
	if err != nil {
		return nil, err
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		return nil, err
	}

	return partitionConsumer, nil
}
