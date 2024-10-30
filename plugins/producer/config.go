package producer

type ProducerConfig struct {	
	TxPoolTopic          string `yaml:"txpool.topic"`         // Topic for mempool transaction data
	BrokerURL            string `yaml:"broker.url"`           // URL of the Cardinal Broker"
	DefaultTopic         string `yaml:"default.topic"`        // "Default topic for Cardinal broker"
	BlockTopic           string `yaml:"block.topic"`          // "Topic for Cardinal block data"
	LogTopic             string `yaml:"logs.topic"`           // "Topic for Cardinal log data"
	TxTopic              string `yaml:"tx.topic"`             // "Topic for Cardinal transaction data"
	ReceiptTopic         string `yaml:"receipt.topic"`        // "Topic for Cardinal receipt data"
	CodeTopic            string `yaml:"code.topic"`           // "Topic for Cardinal contract code"
	StateTopic           string `yaml:"state.topic"`          // "Topic for Cardinal state data"
	StartBlockOverride   uint64 `yaml:"start.block"`          // "The first block to emit"
	ReorgThreshold       int    `yaml:"reorg.threshold"`      // "The number of blocks for clients to support quick reorgs"
	Statsdaddr           string `yaml:"statsd.addr"`          // "UDP address for a statsd endpoint"
	Cloudwatchns         string `yaml:"cloudwatch.namespace"` // "CloudWatch Namespace for cardinal metrics"
	MinActiveProducers   uint   `yaml:"min.producers"`        // "The minimum number of healthy producers for maintenance operations like state trie flush to take place"
}

