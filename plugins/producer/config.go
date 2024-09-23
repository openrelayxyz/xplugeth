package producer

type ProducerConfig struct {	
	TxPoolTopic          string `yaml:"cardinal.txpool.topic"`         // Topic for mempool transaction data
	BrokerURL            string `yaml:"cardinal.broker.url"`           // URL of the Cardinal Broker"
	DefaultTopic         string `yaml:"cardinal.default.topic"`        // "Default topic for Cardinal broker"
	BlockTopic           string `yaml:"cardinal.block.topic"`          // "Topic for Cardinal block data"
	LogTopic             string `yaml:"cardinal.logs.topic"`           // "Topic for Cardinal log data"
	TxTopic              string `yaml:"cardinal.tx.topic"`             // "Topic for Cardinal transaction data"
	ReceiptTopic         string `yaml:"cardinal.receipt.topic"`        // "Topic for Cardinal receipt data"
	CodeTopic            string `yaml:"cardinal.code.topic"`           // "Topic for Cardinal contract code"
	StateTopic           string `yaml:"cardinal.state.topic"`          // "Topic for Cardinal state data"
	StartBlockOverride   uint64 `yaml:"cardinal.start.block"`          // "The first block to emit"
	ReorgThreshold       int    `yaml:"cardinal.reorg.threshold"`      // "The number of blocks for clients to support quick reorgs"
	Statsdaddr           string `yaml:"cardinal.statsd.addr"`          // "UDP address for a statsd endpoint"
	Cloudwatchns         string `yaml:"cardinal.cloudwatch.namespace"` // "CloudWatch Namespace for cardinal metrics"
	MinActiveProducers   uint   `yaml:"cardinal.min.producers"`        // "The minimum number of healthy producers for maintenance operations like state trie flush to take place"
}

