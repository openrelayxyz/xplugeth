package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/openrelayxyz/xplugeth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/node"

	"github.com/Shopify/sarama"
	"github.com/openrelayxyz/cardinal-streams/transports"
)

func GetChainID() (int64, bool) {
	s, ok := xplugeth.GetSingleton[*node.Node]()
	if !ok {
		return 0, false
	}
	var hex hexutil.Uint64
	client := s.Attach()
	defer client.Close()
	client.Call(&hex, "eth_chainId")
	return int64(hex), true
}

func GetTd(hash common.Hash) (*big.Int, error) {
	result := new(big.Int)
	s, ok := xplugeth.GetSingleton[*node.Node]()
	if !ok {
		return nil, errors.New("failed to acqire stack singleton, GetTd")
	}
	var parentBlockJson map[string]json.RawMessage
	client := s.Attach()
	defer client.Close()
	client.Call(&parentBlockJson, "eth_getBlockByHash", hash, false)
	raw, ok := parentBlockJson["totalDifficulty"]
	if !ok {
		chainid, ok := GetChainID()
		if !ok { panic(fmt.Sprintf("could not resolve chain id from within GetTd")) }
		switch chainid {
		case int64(1):
			result.SetString("58750003716598352816469", 10)
		case int64(11155111):
			result.SetString("17000018015853232", 10)
		default:
			result.SetString("1", 10)
		}
		return result, nil
	}
	var td string
	if err := json.Unmarshal(raw, &td); err != nil {
		return nil, err
	}
	if _, ok := result.SetString(td, 0); !ok {
		return nil, errors.New("convert total difficulty string to big int")
	} 
	return result, nil
}

func strPtr(x string) *string {
	return &x
}
var (
	brokers            []string
	config             *sarama.Config
)

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