package agent

import (
	"math/big"
	"strconv"
	"fmt"
	"regexp"
	"runtime"
	"time"
	"sync"
	
	"github.com/openrelayxyz/cardinal-streams/v2/delivery"
	"github.com/openrelayxyz/cardinal-streams/v2/transports"
	"github.com/openrelayxyz/cardinal-types"
	"github.com/openrelayxyz/cardinal-types/metrics"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	beacon "github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	
	log "github.com/inconshreveable/log15"
)

var (
	heightGauge = metrics.NewMajorGauge("/beaconmux/height")
	txRegexp = regexp.MustCompile("b/[0-9a-z]+/b/[0-9a-z]+/t/([0-9a-z]+)")
	vhRegexp = regexp.MustCompile("b/[0-9a-z]+/b/[0-9a-z]+/v/([0-9a-z]+)")
	erRegexp = regexp.MustCompile("b/[0-9a-z]+/b/[0-9a-z]+/er/([0-9a-z]+)")
)

type StreamManager struct{
	consumer transports.Consumer
	client   *rpc.Client
	backendURL string
	sub      types.Subscription
	ready    chan struct{}
	processed uint64
	chainid  int
}

type miniBlock struct {
	Number    hexutil.Uint64 `json:"number"`
	Timestamp hexutil.Uint64 `json:"timestamp"`
	Hash      common.Hash    `json:"hash"`
	Td        hexutil.Big    `json:"totalDifficulty"`
}

func NewStreamManager(brokerParams []transports.BrokerParams, backendURL string, rollbackInSeconds int, terminalDifficulty *big.Int, whitelist map[uint64]types.Hash) (*StreamManager, error) {
	trackedPrefixes := []*regexp.Regexp{
		regexp.MustCompile("b/[0-9a-z]+/b/"),
	}
	
	if err != nil {
		return nil, err
	}
	var block miniBlock
	for i := 0; i < 720 ; i++ { // Retry for up to 1 hour (5 seconds * 720 = 3600 seconds = 1 hour, ignoring Call latency)
		err := sessionClient.Call(&block, "eth_getBlockByNumber", "latest", false);
		if err == nil {
			break
		}
		log.Warn("Failed to get initial block. Retrying.", "err", err, "retries", i)
		time.Sleep(5 * time.Second)
	}
	var chainid hexutil.Uint64
	if err := sessionClient.Call(&chainid, "eth_chainId"); err != nil {
		return nil, err
	}
	processed := uint64(0)
	// lastNum := int64(block.Number)
	// lastHash := types.Hash(block.Hash)
	// lastWeight := new(big.Int).Add(block.Td.ToInt(), big.NewInt(int64(block.Number)))
	// resumption, err := transports.ResumptionForTimestamp(brokerParams, int64(int(block.Timestamp) - rollbackInSeconds) * 1000)
	// if err != nil {
	// 	log.Warn("Could not generate resumption token", "err", err)
	// }
	_, err = transports.ResumptionForTimestamp(brokerParams, int64(int(block.Timestamp) - rollbackInSeconds) * 1000)
	if err != nil {
		log.Warn("Could not generate resumption token", "err", err)
	}

	var consumer transports.Consumer
	// It should be safe to calculate resumption weight this way. The producer
	// won't start publishing until the merge block, at which point it will set
	// the weight to Td + block Number. If we start up before the merge,
	var emptyHash types.Hash
	var emptyResumption []byte
	consumer, err = transports.ResolveMuxConsumer(brokerParams, emptyResumption, &delivery.ConsumerConfig{
		LastEmittedNum: 0,
		LastHash: emptyHash,
		LastWeight: new(big.Int),
		ReorgThreshold: 128,
		TrackedPrefixes: trackedPrefixes,
		Whitelist: whitelist,
	})
	if err != nil { return nil, err }
	return &StreamManager{
		consumer: consumer,
		client: sessionClient,
		ready: make(chan struct{}, 1),
		chainid: int(chainid),
		backendURL: backendURL,
		processed: processed,
	}, nil
}

func encodeParams(ed beacon.ExecutableData) map[string]interface{} {
	// ParentHash    common.Hash         `json:"parentHash"    gencodec:"required"`
	// 	FeeRecipient  common.Address      `json:"feeRecipient"  gencodec:"required"`
	// 	StateRoot     common.Hash         `json:"stateRoot"     gencodec:"required"`
	// 	ReceiptsRoot  common.Hash         `json:"receiptsRoot"  gencodec:"required"`
	// 	LogsBloom     hexutil.Bytes       `json:"logsBloom"     gencodec:"required"`
	// 	Random        common.Hash         `json:"prevRandao"    gencodec:"required"`
	// 	Number        hexutil.Uint64      `json:"blockNumber"   gencodec:"required"`
	// 	GasLimit      hexutil.Uint64      `json:"gasLimit"      gencodec:"required"`
	// 	GasUsed       hexutil.Uint64      `json:"gasUsed"       gencodec:"required"`
	// 	Timestamp     hexutil.Uint64      `json:"timestamp"     gencodec:"required"`
	// 	ExtraData     hexutil.Bytes       `json:"extraData"     gencodec:"required"`
	// 	BaseFeePerGas *hexutil.Big        `json:"baseFeePerGas" gencodec:"required"`
	// 	BlockHash     common.Hash         `json:"blockHash"     gencodec:"required"`
	// 	Transactions  []hexutil.Bytes     `json:"transactions"  gencodec:"required"`
	// 	Withdrawals   []*types.Withdrawal `json:"withdrawals"`
	result := make(map[string]interface{})
	result["parentHash"] = common.Hash(ed.ParentHash)
	result["feeRecipient"] = common.Address(ed.FeeRecipient)
	result["stateRoot"] = common.Hash(ed.StateRoot)
	result["receiptsRoot"] = common.Hash(ed.ReceiptsRoot)
	result["logsBloom"] = hexutil.Bytes(ed.LogsBloom)
	result["prevRandao"] = common.Hash(ed.Random)
	result["blockNumber"] = hexutil.Uint64(ed.Number)
	result["gasLimit"] = hexutil.Uint64(ed.GasLimit)
	result["gasUsed"] = hexutil.Uint64(ed.GasUsed)
	result["timestamp"] = hexutil.Uint64(ed.Timestamp)
	result["extraData"] = hexutil.Bytes(ed.ExtraData)
	result["baseFeePerGas"] = (*hexutil.Big)(ed.BaseFeePerGas)
	result["blockHash"] = common.Hash(ed.BlockHash)
	txs := make([]hexutil.Bytes, len(ed.Transactions))
	for i, t := range ed.Transactions {
		txs[i] = hexutil.Bytes(t)
	}
	result["transactions"] = txs
	if ed.Withdrawals != nil {
		result["withdrawals"] = ed.Withdrawals
	}
	if ed.BlobGasUsed != nil {
		result["blobGasUsed"] = hexutil.Uint64(*ed.BlobGasUsed)
	}
	if ed.ExcessBlobGas != nil {
		result["excessBlobGas"] = hexutil.Uint64(*ed.ExcessBlobGas)
	}
	return result

}

func (m *StreamManager) Start() <-chan error {
	errCh := make(chan error, 1)
	log.Info("Starting agent")
	if m.sub != nil {
		errCh <- fmt.Errorf("already started")
		return errCh
	}
	var lock sync.Mutex
	ch := make(chan *delivery.ChainUpdate)
	m.sub = m.consumer.Subscribe(ch)
	go func() {
		<-m.consumer.Ready()
		runtime.Gosched() // Give the processing goroutine a chance to start processing updates
		lock.Lock() // Wait for the processing goroutine a chance to finish processing the current update
		m.ready <- struct{}{}
		lock.Unlock()
	}()
	go func() {
		for {
			log.Debug("Waiting for message")
			processed := uint64(0)
			select {
			case update := <-ch:
				lock.Lock()
				// start := time.Now()
				forkChoiceMethod := "engine_forkchoiceUpdatedV1"
				added := update.Added()
				var latestHash, safeHash, finalHash common.Hash
				for _, pb := range added {
					log.Error("inside of pb for loop")
					var params beacon.ExecutableData
					if err := rlp.DecodeBytes(pb.Values[fmt.Sprintf("b/%x/b/%x/h", m.chainid, pb.Hash.Bytes())], &params); err != nil {
						log.Warn("Failed to rlp Decode block", "number", pb.Number, "err", err)
						continue
					}
					txs := make(map[int][]byte)
					versionedHashesMap := make(map[int]common.Hash)
					executionRequestsMap := make(map[int][]byte)
					for k, v := range pb.Values {
						log.Error("inside of the pb.Values range")
						switch {
						case txRegexp.MatchString(k):
							log.Error("inside txRegexp case")
							parts := txRegexp.FindSubmatch([]byte(k))
							txIndex, _ := strconv.ParseInt(string(parts[1]), 16, 64)
							txs[int(txIndex)] = v
						case vhRegexp.MatchString(k):
							log.Error("inside vhRegexp case")
							parts := vhRegexp.FindSubmatch([]byte(k))
							idx, _ := strconv.ParseInt(string(parts[1]), 16, 64)
							versionedHashesMap[int(idx)] = common.BytesToHash(v)
						case erRegexp.MatchString(k):
							log.Error("inside erRegexp case")
							parts := erRegexp.FindSubmatch([]byte(k))
							idx, _ := strconv.ParseInt(string(parts[1]), 16, 64)
							executionRequestsMap[int(idx)] = v
						}
					}
					versionedHashes := make([]common.Hash, len(versionedHashesMap))
					for i := 0; i < len(versionedHashesMap); i++ {
						versionedHashes[i] = versionedHashesMap[i]
					}
					params.Transactions = make([][]byte, len(txs))
					for i := 0 ; i < len(txs); i++ {
						v, ok := txs[i]
						if !ok {
							log.Warn("Value missing from transactions list", "block", pb.Hash, "txindex", i)
						}
						params.Transactions[i] = v
					}
					ep := encodeParams(params)
					isV4 := false
					var executionRequests []hexutil.Bytes
					if len(executionRequestsMap) > 0 {
						isV4 = true
						if len(executionRequestsMap[0]) == 0 {
							// We set this to empty just to signal that it's a v4 block, but there are no requests, so make an empty list
							executionRequests = make([]hexutil.Bytes, 0)
						} else {
							executionRequests = make([]hexutil.Bytes, len(executionRequestsMap))
							for k, v := range executionRequestsMap {
								executionRequests[k] = v
							}
						}
					}
					log.Info("Calling engine_newPayload", "block", pb.Number)
					var res interface{}
					if isV4 {
						beaconRootBytes := pb.Values[fmt.Sprintf("b/%x/b/%x/br", m.chainid, params.BlockHash.Bytes())]
						beaconRoot := common.BytesToHash(beaconRootBytes)
						forkChoiceMethod = "engine_forkchoiceUpdatedV3"
						log.Debug("Calling engine_newPayloadV4", "ep", ep, "vh", versionedHashes, "br", beaconRoot, "er", executionRequests)
						if err := m.client.Call(&res, "engine_newPayloadV4", ep, versionedHashes, beaconRoot, executionRequests); err != nil {
							log.Warn("Error sending payload", "block", pb.Number, "err", err)
							errCh <- err
							lock.Unlock()
							return
						}
					} else if beaconRootBytes, ok := pb.Values[fmt.Sprintf("b/%x/b/%x/br", m.chainid, params.BlockHash.Bytes())]; ok {
						beaconRoot := common.BytesToHash(beaconRootBytes)
						forkChoiceMethod = "engine_forkchoiceUpdatedV3"
						log.Debug("Calling engine_newPayloadV3", "ep", ep, "vh", versionedHashes, "br", beaconRoot)
						if err := m.client.Call(&res, "engine_newPayloadV3", ep, versionedHashes, beaconRoot); err != nil {
							log.Warn("Error sending payload", "block", pb.Number, "err", err)
							errCh <- err
							lock.Unlock()
							return
						}
					} else {
						delete(ep, "excessBlobGas")
						delete(ep, "blobGasUsed")
						log.Debug("Calling engine_newPayloadV2", "ep", ep)
						if err := m.client.Call(&res, "engine_newPayloadV2", ep); err != nil {
							if err.Error() == "Invalid parameters" {
								delete(ep, "withdrawals")
								log.Debug("Calling engine_newPayloadV2", "ep", ep)
								if err := m.client.Call(&res, "engine_newPayloadV2", ep); err != nil {
									log.Warn("Error sending payload", "block", pb.Number, "err", err)
									errCh <- err
									lock.Unlock()
									return
								}
							} else {
								log.Warn("Error sending payload", "block", pb.Number, "err", err)
								errCh <- err
								lock.Unlock()
								return
							}
						}
					}
					heightGauge.Update(pb.Number)
					processed++
					pb.Done()
					latestHash = common.Hash(pb.Hash)
					if v, ok := pb.Values[fmt.Sprintf("b/%x/b/%x/s", m.chainid, params.BlockHash.Bytes())]; ok {
						safeHash = common.BytesToHash(v)
					} else {
						safeHash = latestHash
					}
					if v, ok := pb.Values[fmt.Sprintf("b/%x/b/%x/f", m.chainid, params.BlockHash.Bytes())]; ok {
						finalHash = common.BytesToHash(v)
					}
				}
				var fcr beacon.ForkChoiceResponse
				log.Info("Calling engine_forkchoiceUpdated", "block", latestHash)
				if err := m.client.Call(&fcr, forkChoiceMethod, beacon.ForkchoiceStateV1{
					HeadBlockHash: latestHash,
					SafeBlockHash: safeHash,
					FinalizedBlockHash: finalHash,
				}, nil); err != nil {
					log.Warn("Error sending forkchoice payload", "hash", latestHash, "err", err)
					errCh <- err
					lock.Unlock()
					return
				}
				log.Error("payload status from the agent", "status", fcr.PayloadStatus.Status)
				if fcr.PayloadStatus.Status == beacon.VALID {
					log.Error("Inside condition", "status", fcr.PayloadStatus.Status)
					m.processed += processed
				}				
			}
			lock.Unlock()
		}
	}()
	if err := m.consumer.Start(); err != nil {
		errCh <- err
	}
	return errCh
}

func (m *StreamManager) Ready() chan struct{}{
	return m.ready
}

func (m *StreamManager) ChainID() int {
	return m.chainid
}

func (m *StreamManager) Close() {
	m.sub.Unsubscribe()
	m.consumer.Close()
}

func (m *StreamManager) API() *api {
	return &api{m.consumer}
}

func (m *StreamManager) Processed() uint64 {
	return m.processed
}

type api struct{
	consumer transports.Consumer
}

func (a *api) WhyNotReady(hash types.Hash) string {
	return a.consumer.WhyNotReady(hash)
}