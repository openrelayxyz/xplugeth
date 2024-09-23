package producer

type ProducerConfig struct {	
	txPoolTopic          string `yaml:"cardinal.txpool.topic"`         // Topic for mempool transaction data
	brokerURL            string `yaml:"cardinal.broker.url"`           // URL of the Cardinal Broker"
	defaultTopic         string `yaml:"cardinal.default.topic"`        // "Default topic for Cardinal broker"
	blockTopic           string `yaml:"cardinal.block.topic"`          // "Topic for Cardinal block data"
	logTopic             string `yaml:"cardinal.logs.topic"`           // "Topic for Cardinal log data"
	txTopic              string `yaml:"cardinal.tx.topic"`             // "Topic for Cardinal transaction data"
	receiptTopic         string `yaml:"cardinal.receipt.topic"`        // "Topic for Cardinal receipt data"
	codeTopic            string `yaml:"cardinal.code.topic"`           // "Topic for Cardinal contract code"
	stateTopic           string `yaml:"cardinal.state.topic"`          // "Topic for Cardinal state data"
	startBlockOverride   uint64 `yaml:"cardinal.start.block"`          // "The first block to emit"
	reorgThreshold       int    `yaml:"cardinal.reorg.threshold"`      // "The number of blocks for clients to support quick reorgs"
	statsdaddr           string `yaml:"cardinal.statsd.addr"`          // "UDP address for a statsd endpoint"
	cloudwatchns         string  `yaml:"cardinal.cloudwatch.namespace"` // "CloudWatch Namespace for cardinal metrics"
	minActiveProducers   uint   `yaml:"cardinal.min.producers"`        // "The minimum number of healthy producers for maintenance operations like state trie flush to take place"
}

