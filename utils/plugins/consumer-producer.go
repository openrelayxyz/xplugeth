package plugins

import (

	"fmt"
	"strings"
	"github.com/Shopify/sarama"
	"github.com/openrelayxyz/cardinal-streams/transports"
)

var (
	brokers            []string
	config             *sarama.Config
)

func strPtr(x string) *string {
	return &x
}

func CreateProducer(broker, topic string) (sarama.AsyncProducer, error) {

	brokers, config = transports.ParseKafkaURL(strings.TrimPrefix(broker, "kafka://"))
	configEntries := make(map[string]*string)
	configEntries["retention.ms"] = strPtr("1800000")

	if err := transports.CreateTopicIfDoesNotExist(strings.TrimPrefix(broker, "kafka://"), topic, 1, configEntries); err != nil {
		panic(fmt.Sprintf("Could not create topic %v on broker %v: %v", topic, broker, err.Error()))
	}

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(fmt.Sprintf("Could not setup producer, peer manager plugin: %v", err.Error()))
	}

	return producer, nil
}

func CreateConsumer(broker, topic string) (sarama.PartitionConsumer, error) {

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, err
	}

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		return nil, err
	}

	return partitionConsumer, nil
}