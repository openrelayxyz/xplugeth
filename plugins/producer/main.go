package producer

import (
	"fmt"
	"context"
	"math/big"
	"net"
	"time"
	"strings"
	"sync"

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	
	ctypes "github.com/openrelayxyz/cardinal-types"
	"github.com/openrelayxyz/cardinal-streams/delivery"
	"github.com/openrelayxyz/cardinal-streams/transports"
	"github.com/openrelayxyz/cardinal-types/metrics"
	
	"github.com/Shopify/sarama"
	"github.com/savaki/cloudmetrics"
	"github.com/pubnub/go-metrics-statsd"
)

// the imports below are bringing in other plugins which have to be present for the producer to function properly
import (
	_ "github.com/openrelayxyz/xplugeth/plugins/blockupdates"
	_ "github.com/openrelayxyz/xplugeth/plugins/merge"
)

type cardinalProducerModule struct {
	stack   *node.Node
	backend types.Backend
}

type externalBlockUpdates interface {
	BlockUpdatesByNumber(int64) (*gtypes.Block, *big.Int, gtypes.Receipts, map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte, error)
}

type externalAddBlock interface {
	CardinalAddBlockHook(int64, ctypes.Hash, ctypes.Hash, *big.Int, map[string][]byte, map[string]struct{})
}

type externalTestPlugin interface {
	ExternUpdatesTest() string
}

type externalStreamSchema interface {
	UpdateStreamsSchema(map[string]string)
}

func init() {
	xplugeth.RegisterModule[cardinalProducerModule]()
	xplugeth.RegisterHook[externalBlockUpdates]()
	xplugeth.RegisterHook[externalAddBlock]()
	xplugeth.RegisterHook[externalStreamSchema]()
	xplugeth.RegisterHook[externalTestPlugin]()
}

func (*cardinalProducerModule) ExternProducerTest() string {
	return "calling from caridnal producer"
}

var (
	cfg *ProducerConfig
	ready sync.WaitGroup
	backend types.Backend
	stack  *node.Node
	config *params.ChainConfig
	chainid int64
	producer transports.Producer
	startBlock uint64
	pendingReorgs map[common.Hash]func()
	gethHeightGauge = metrics.NewMajorGauge("/geth/height")
	gethPeersGauge = metrics.NewMajorGauge("/geth/peers")
	masterHeightGauge = metrics.NewMajorGauge("/master/height")
	blockAgeTimer = metrics.NewMajorTimer("/geth/age")
	blockUpdatesByNumber func(number int64)(*gtypes.Block, *big.Int, gtypes.Receipts, map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte, error)

	addBlockHook func(number int64, hash, parent ctypes.Hash, weight *big.Int, updates map[string][]byte, deletes map[string]struct{})
)

func strPtr(x string) *string {
	return &x
}

func (*cardinalProducerModule) InitializeNode(s *node.Node, b types.Backend) {

	var ok bool
	cfg, ok = xplugeth.GetConfig[ProducerConfig]("producer")
	if !ok {
		cfg = &ProducerConfig{ ReorgThreshold:128 }
		log.Warn("no config found, producer plugin, all values set to default")
	}

	backend = b
	stack = s
	ready.Add(1)
	defer ready.Done()
	config = b.ChainConfig()
	chainid = config.ChainID.Int64()
	pendingReorgs = make(map[common.Hash]func())

	for _, updater := range xplugeth.GetModules[externalBlockUpdates]() {
		blockUpdatesByNumber = updater.BlockUpdatesByNumber
	}

	addBlockFns := []func(int64, ctypes.Hash, ctypes.Hash, *big.Int, map[string][]byte, map[string]struct{}){}
	for _, extern := range xplugeth.GetModules[externalAddBlock]() {
		addBlockFns = append(addBlockFns, extern.CardinalAddBlockHook)
	}

	addBlockHook = func(number int64, hash, parent ctypes.Hash, td *big.Int, updates map[string][]byte, deletes map[string]struct{}) {
		for _, fn := range addBlockFns {
			fn(number, hash, parent, td, updates, deletes)
		}
	}
	

	if cfg.DefaultTopic == "" { cfg.DefaultTopic = fmt.Sprintf("cardinal-%v", chainid) }
	if cfg.BlockTopic == "" { cfg.BlockTopic = fmt.Sprintf("%v-block", cfg.DefaultTopic) }
	if cfg.LogTopic == "" { cfg.LogTopic = fmt.Sprintf("%v-logs", cfg.DefaultTopic) }
	if cfg.TxTopic == "" { cfg.TxTopic = fmt.Sprintf("%v-tx", cfg.DefaultTopic) }
	if cfg.ReceiptTopic == "" { cfg.ReceiptTopic = fmt.Sprintf("%v-receipt", cfg.DefaultTopic) }
	if cfg.CodeTopic == "" { cfg.CodeTopic = fmt.Sprintf("%v-code", cfg.DefaultTopic) }
	if cfg.StateTopic == "" { cfg.StateTopic = fmt.Sprintf("%v-state", cfg.DefaultTopic) }
	var err error
	brokers := []transports.ProducerBrokerParams{
		{
			URL: "ws://0.0.0.0:8555",
		},
	}
	schema := map[string]string{
		fmt.Sprintf("c/%x/a/", chainid): cfg.StateTopic,
		fmt.Sprintf("c/%x/s", chainid): cfg.StateTopic,
		fmt.Sprintf("c/%x/c/", chainid): cfg.CodeTopic,
		fmt.Sprintf("c/%x/b/[0-9a-z]+/h", chainid): cfg.BlockTopic,
		fmt.Sprintf("c/%x/b/[0-9a-z]+/d", chainid): cfg.BlockTopic,
		fmt.Sprintf("c/%x/b/[0-9a-z]+/w", chainid): cfg.BlockTopic,
		fmt.Sprintf("c/%x/b/[0-9a-z]+/u/", chainid): cfg.BlockTopic,
		fmt.Sprintf("c/%x/n/", chainid): cfg.BlockTopic,
		fmt.Sprintf("c/%x/b/[0-9a-z]+/t/", chainid): cfg.TxTopic,
		fmt.Sprintf("c/%x/b/[0-9a-z]+/r/", chainid): cfg.ReceiptTopic,
		fmt.Sprintf("c/%x/b/[0-9a-z]+/l/", chainid): cfg.LogTopic,
	}

	// Let plugins add schema updates for any values they will provide.
	for _, extern := range xplugeth.GetModules[externalStreamSchema]() {
		extern.UpdateStreamsSchema(schema)
	}

	if strings.HasPrefix(cfg.BrokerURL, "kafka://") {
		brokers = append(brokers, transports.ProducerBrokerParams{
			URL: cfg.BrokerURL,
			DefaultTopic: cfg.DefaultTopic,
			Schema: schema,
		})
	} else if strings.HasPrefix(cfg.BrokerURL, "file://") {
		brokers = append(brokers, transports.ProducerBrokerParams{
			URL: cfg.BrokerURL,
		})
	}
	log.Info("Producing to brokers", "brokers", brokers, "burl", cfg.BrokerURL)
	producer, err = transports.ResolveMuxProducer(
		brokers,
		&resumer{},
	)
	if err != nil { panic(err.Error()) }
	if minap := cfg.MinActiveProducers; minap > 0 {
		go func() {
			client := stack.Attach()
			for client == nil {
				time.Sleep(500 * time.Second)
				client = stack.Attach()
			}
			t := time.NewTicker(5 * time.Second)
			enabled := true
			for range t.C {
				pc := producer.ProducerCount(time.Minute)
				if  pc >= minap && !enabled {
					// If there are adequate healthy producers, the flush interval should be 1 hour
					var res any
					client.Call(&res, "debug_setTrieFlushInterval", "1h") // We're not error checking this in case we're on a node that doesn't support this method
					enabled = true
				} else if pc < minap && enabled {
					// If there aren't enough healthy producers, set the flush interval very high
					var res any
					client.Call(&res, "debug_setTrieFlushInterval", "72h") // We're not error checking this in case we're on a node that doesn't support this method
					enabled = false
				}
				// I feel like it's useful to note the behavior of debug_setTrieFlushInterval here.
				// A trie flush interval of 1 hour does not mean that the trie will be flushed every hour.
				// It means that the trie will be flushed after 1 hour of time spent processing blocks.
				// The intent is that if you have an unclean shutdown and a node has to start back
				// up from the last trie flush, it should take no more than 1 hour of processing to
				// catch back up. If it takes 500ms to process a block, and blocks come out every 12
				// seconds, it will take ~24 hours to accumulate an hour of block processing, so the
				// trie flush will only happen once every 24 hours. Setting the trie flush interval to
				// 72 hours then means that a trie flush may only happen every couple of months, but that
				// an unclean restart will take less than 72 hours to resume.
				//
				// Needless to say, you don't want to find yourself in that position, so it's advised to
				// make sure you have multiple healthy masters if you are setting the --cardinal.min.producers
				// flag
			}
		}()
	}
	if cfg.BrokerURL != "" {
		go func() {
			t := time.NewTicker(time.Second * 30)
			defer t.Stop()
			for range t.C {
				gethPeersGauge.Update(int64(stack.Server().PeerCount()))
			}
		}()
		if cfg.Statsdaddr != "" {
			udpAddr, err := net.ResolveUDPAddr("udp", cfg.Statsdaddr)
			if err != nil {
				log.Error("Invalid Address. Statsd will not be configured.", "error", err.Error())
			}
			go statsd.StatsD(
				metrics.MajorRegistry,
				20 * time.Second,
				"cardinal.geth.master",
				udpAddr,
			)
		}
		if cfg.Cloudwatchns != "" {
			go cloudmetrics.Publish(metrics.MajorRegistry,
				cfg.Cloudwatchns,
				cloudmetrics.Dimensions("chainid", fmt.Sprintf("%v", chainid)),
				cloudmetrics.Interval(30 * time.Second),
			)
		}
		if cfg.StartBlockOverride > 0 {
			startBlock = cfg.StartBlockOverride
		} else {
			v, err := producer.LatestBlockFromFeed()
			if err != nil {
				log.Error("Error getting start block", "err", err)
			} else {
				if v > 128 {
					startBlock = uint64(v) - 128
					log.Info("Setting start block from producer", "block", startBlock)
				}
			}
		}
		if cfg.TxPoolTopic != "" {
			go func() {
				// TODO: we should probably do something within Cardinal streams to
				// generalize this so it's not Kafka specific and can work with other
				// transports.
				ch := make(chan core.NewTxsEvent, 1000)
				sub := b.SubscribeNewTxsEvent(ch)
				brokers, config := transports.ParseKafkaURL(strings.TrimPrefix(cfg.BrokerURL, "kafka://"))
				configEntries := make(map[string]*string)
				configEntries["retention.ms"] = strPtr("3600000")
				if err := transports.CreateTopicIfDoesNotExist(strings.TrimPrefix(cfg.BrokerURL, "kafka://"), cfg.TxPoolTopic, 0, configEntries); err != nil {
					panic(fmt.Sprintf("Could not create topic %v on broker %v: %v", cfg.TxPoolTopic, cfg.BrokerURL, err.Error()))
				}
				// This is about twice the size of the largest possible transaction if
				// all gas in a block were zero bytes in a transaction's data. It should
				// be very rare for messages to even approach this size.
				config.Producer.MaxMessageBytes = 10000024
				producer, err := sarama.NewAsyncProducer(brokers, config)
				if err != nil {
					panic(fmt.Sprintf("Could not setup producer: %v", err.Error()))
				}
				for {
					select {
					case txEvent := <-ch:
						for _, tx := range txEvent.Txs {
							// Switch from MarshalBinary to RLP encoding to match EtherCattle's legacy format for txpool transactions
							txdata, err := rlp.EncodeToBytes(tx)
							if err == nil {
								select {
								case producer.Input() <- &sarama.ProducerMessage{Topic: cfg.TxPoolTopic, Value: sarama.ByteEncoder(txdata)}:
								case err := <-producer.Errors():
									log.Error("Error emitting: %v", "err", err.Error())
								}
							} else {
								log.Warn("Error RLP encoding transactions", "err", err)
							}
						}
					case err := <-sub.Err():
						log.Error("Error processing event transactions", "error", err)
						close(ch)
						sub.Unsubscribe()
						return
					}
				}
			}()
		}
	}
	log.Error("Cardinal EVM plugin initialized")

	for _, updater := range xplugeth.GetModules[externalTestPlugin]() {
		log.Error("from block updater", "response", updater.ExternUpdatesTest())
		
	}

}

type receiptMeta struct {
	ContractAddress common.Address
	CumulativeGasUsed uint64
	GasUsed uint64
	LogsBloom []byte
	Status uint64
	LogCount uint
	LogOffset uint
}

func BUPreReorg(common common.Hash, oldChain []common.Hash, newChain []common.Hash) {
	block, err := backend.BlockByHash(context.Background(), common)
	if err != nil {
		log.Error("Could not get block for reorg", "hash", common, "err", err)
		return
	}
	
	if len(oldChain) > cfg.ReorgThreshold && len(newChain) > 0 {
		pendingReorgs[common], err = producer.Reorg(int64(block.NumberU64()), ctypes.Hash(common))
		if err != nil {
			log.Error("Could not start producer reorg", "block", common, "num", block.NumberU64(), "err", err)
		}
	}
}

type resumer struct {}

func (*resumer) GetBlock(ctx context.Context, number uint64) (*delivery.PendingBatch) {
	block, td, receipts, destructs, accounts, storage, code, err := blockUpdatesByNumber(int64(number))
	if block == nil {
		log.Warn("Error retrieving block", "number", number, "err", err)
		return nil
	}
	if destructs == nil || accounts == nil || storage == nil || code == nil {
		destructs, accounts, storage, code, err = stateTrieUpdatesByNumber(int64(number))
		if err != nil {
			log.Warn("Could not retrieve block state", "err", err)
		}
	}

	hash := block.Hash()
	weight, updates, deletes, _, batchUpdates := getUpdates(block, td, receipts, destructs, accounts, storage, code)
	// Since we're just sending a single PendingBatch, we need to merge in
	// updates. Once we add support for plugins altering the above, we may
	// need to handle deletes in batchUpdates.
	for _, b := range batchUpdates {
		for k, v := range b {
			updates[k] = v
		}
	}
	return &delivery.PendingBatch{
		Number: int64(number),
		Weight: weight,
		ParentHash: ctypes.Hash(block.ParentHash()),
		Hash: ctypes.Hash(hash),
		Values: updates,
		Deletes: deletes,
	}
}

func (r *resumer) BlocksFrom(ctx context.Context, number uint64, hash ctypes.Hash) (chan *delivery.PendingBatch, error) {
	if blockUpdatesByNumber == nil {
		return nil, fmt.Errorf("cannot retrieve old block updates")
	}
	ch := make(chan *delivery.PendingBatch)
	go func() {
		reset := false
		defer close(ch)
		for i := number; ; i++ {
			if pb := r.GetBlock(ctx, i); pb != nil {
				if pb.Number == int64(number) && (pb.Hash != hash) && !reset {
					reset = true
					if int(i) < cfg.ReorgThreshold {
						i = 0
					} else {
						i -= uint64(cfg.ReorgThreshold)
					}
					continue
				}
				select {
				case <-ctx.Done():
					return
				case ch <- pb:
				}
			} else {
				return
			}
		}
	}()
	return ch, nil
}

func BUPostReorg(common common.Hash, oldChain []common.Hash, newChain []common.Hash) {
	if done, ok := pendingReorgs[common]; ok {
		done()
		delete(pendingReorgs, common)
	}
}

func getUpdates(block *gtypes.Block, td *big.Int, receipts gtypes.Receipts, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, code map[common.Hash][]byte) (*big.Int, map[string][]byte, map[string]struct{}, map[string]ctypes.Hash, map[ctypes.Hash]map[string][]byte) {
	hash := block.Hash()
	headerBytes, _ := rlp.EncodeToBytes(block.Header())
	updates := map[string][]byte{
		fmt.Sprintf("c/%x/b/%x/h", chainid, hash.Bytes()): headerBytes,
		fmt.Sprintf("c/%x/b/%x/d", chainid, hash.Bytes()): td.Bytes(),
		fmt.Sprintf("c/%x/n/%x", chainid, block.Number().Int64()): hash[:],
	}
	if block.Withdrawals().Len() > 0 {
		withdrawalsBytes, _ := rlp.EncodeToBytes(block.Withdrawals())
		updates[fmt.Sprintf("c/%x/b/%x/w", chainid, hash.Bytes())] = withdrawalsBytes
	}
	for i, tx := range block.Transactions() {
		updates[fmt.Sprintf("c/%x/b/%x/t/%x", chainid, hash.Bytes(), i)], _ = tx.MarshalBinary()
		rmeta := receiptMeta{
			ContractAddress: common.Address(receipts[i].ContractAddress),
			CumulativeGasUsed: receipts[i].CumulativeGasUsed,
			GasUsed: receipts[i].GasUsed,
			LogsBloom: receipts[i].Bloom.Bytes(),
			Status: receipts[i].Status,
			LogCount: uint(len(receipts[i].Logs)),
		}
		if rmeta.LogCount > 0 {
			rmeta.LogOffset = receipts[i].Logs[0].Index
		}
		updates[fmt.Sprintf("c/%x/b/%x/r/%x", chainid, hash.Bytes(), i)], _ = rlp.EncodeToBytes(rmeta)
		for _, logRecord := range receipts[i].Logs {
			updates[fmt.Sprintf("c/%x/b/%x/l/%x/%x", chainid, hash.Bytes(), i, logRecord.Index)], _ = rlp.EncodeToBytes(logRecord)
		}
	}
	for hashedAddr, acctRLP := range accounts {
		updates[fmt.Sprintf("c/%x/a/%x/d", chainid, hashedAddr.Bytes())] = acctRLP
	}
	for codeHash, codeBytes := range code {
		updates[fmt.Sprintf("c/%x/c/%x", chainid, codeHash.Bytes())] = codeBytes
	}
	deletes := make(map[string]struct{})
	for hashedAddr := range destructs {
		deletes[fmt.Sprintf("c/%x/a/%x", chainid, hashedAddr.Bytes())] = struct{}{}
	}
	batches := map[string]ctypes.Hash{
		fmt.Sprintf("c/%x/s", chainid): ctypes.BigToHash(block.Number()),
	}
	if len(block.Uncles()) > 0 {
		// If uncles == 0, we can figure that out from the hash without having to
		// send an empty list across the wire
		for i, uncle := range block.Uncles() {
			updates[fmt.Sprintf("c/%x/b/%x/u/%x", chainid, hash.Bytes(), i)], _ = rlp.EncodeToBytes(uncle)
		}
	}
	batchUpdates := map[ctypes.Hash]map[string][]byte{
		ctypes.BigToHash(block.Number()): make(map[string][]byte),
	}
	for addrHash, updates := range storage {
		for k, v := range updates {
			batchUpdates[ctypes.BigToHash(block.Number())][fmt.Sprintf("c/%x/a/%x/s/%x", chainid, addrHash.Bytes(), k.Bytes())] = v
		}
	}

	weight := new(big.Int).Set(td)
	addBlockHook(block.Number().Int64(), ctypes.Hash(hash), ctypes.Hash(block.ParentHash()), weight, updates, deletes)
	return weight, updates, deletes, batches, batchUpdates
}

func (*cardinalProducerModule) BlockUpdates(block *gtypes.Block, td *big.Int, receipts gtypes.Receipts, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, code map[common.Hash][]byte) {
	if producer == nil {
		panic("Unknown broker. Please set --cardinal.broker.url")
	}
	ready.Wait()
	if block.NumberU64() < startBlock {
		log.Debug("Skipping block production", "current", block.NumberU64(), "start", startBlock)
		return
	}
	hash := block.Hash()
	weight, updates, deletes, batches, batchUpdates := getUpdates(block, td, receipts, destructs, accounts, storage, code)
	log.Info("Producing block to cardinal-streams", "hash", hash, "number", block.NumberU64())
	gethHeightGauge.Update(block.Number().Int64())
	masterHeightGauge.Update(block.Number().Int64())
	producer.SetHealth(time.Since(time.Unix(int64(block.Time()), 0)) < 2 * time.Minute)
	blockAgeTimer.UpdateSince(time.Unix(int64(block.Time()), 0))
	if err := producer.AddBlock(
		block.Number().Int64(),
		ctypes.Hash(hash),
		ctypes.Hash(block.ParentHash()),
		weight,
		updates,
		deletes,
		batches,
	); err != nil {
		log.Error("Failed to send block", "block", hash, "err", err)
		panic(err.Error())
	}
	for batchid, update := range batchUpdates {
		if err := producer.SendBatch(batchid, []string{}, update); err != nil {
			log.Error("Failed to send state batch", "block", hash, "err", err)
			return
		}
	}
}

func (*cardinalProducerModule) PreTrieCommit(common.Hash) {
	if producer != nil {
		producer.SetHealth(false)
	}
}
func (*cardinalProducerModule) PostTrieCommit(common.Hash) {
	if producer != nil {
		producer.SetHealth(true)
	}
}


type cardinalAPI struct {
	stack   *node.Node
	blockUpdatesByNumber func(int64) (*gtypes.Block, *big.Int, gtypes.Receipts, map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte, error)
	producer *cardinalProducerModule
}

func (api *cardinalAPI) ReproduceBlocks(start rpc.BlockNumber, end *rpc.BlockNumber) (bool, error) {
	client := api.stack.Attach()
	
	var currentBlock int64
	client.Call(&currentBlock, "eth_blockNumber")
	fromBlock := start.Int64()
	if fromBlock < 0 {
		fromBlock = currentBlock
	}
	var toBlock int64
	if end == nil {
		toBlock = fromBlock
	} else {
		toBlock = end.Int64()
	}
	if toBlock < 0 {
		toBlock = int64(currentBlock)
	}
	oldStartBlock := startBlock
	startBlock = 0
	for i := fromBlock; i <= toBlock; i++ {
		block, td, receipts, destructs, accounts, storage, code, err := api.blockUpdatesByNumber(i)
		if err != nil {
			return false, err
		}
		api.producer.BlockUpdates(block, td, receipts, destructs, accounts, storage, code)
	}
	startBlock = oldStartBlock
	return true, nil
}

func (c *cardinalProducerModule) GetAPIs(stack *node.Node, backend types.Backend) []rpc.API {
	var v func(int64) (*gtypes.Block, *big.Int, gtypes.Receipts, map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte, error)
	for _, extern := range xplugeth.GetModules[externalBlockUpdates]() {
		v = extern.BlockUpdatesByNumber
	}
	if v == nil { log.Warn("Could not load BlockUpdatesByNumber. cardinal_reproduceBlocks will not be available") }
	return []rpc.API{
		{
			Namespace:	"cardinal",
			Version:		"1.0",
			Service:		&cardinalAPI{
				stack,
				v,
				c,
			},
			Public:		true,
		},
		{
			Namespace: "cardinalmetrics",
			Version: "1.0",
			Service: &metrics.MetricsAPI{},
			Public: true,
		},
	}
}

func (c *cardinalAPI) CardinalTest(context.Context) string {
	var notNill bool
	if c.stack != nil {
		notNill = true
	}
	return fmt.Sprintf("Reprting from producer, the stack object is not nil: %v", notNill)
}
