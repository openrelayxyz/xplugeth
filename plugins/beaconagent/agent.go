package beaconagent

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"time"

	beacon "github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/openrelayxyz/cardinal-streams/delivery"
	"github.com/openrelayxyz/cardinal-streams/transports"
	types "github.com/openrelayxyz/cardinal-types"
	"github.com/openrelayxyz/cardinal-types/metrics"
)

var (
	heightGauge = metrics.NewMajorGauge("/beaconmux/height")
	txRegexp    = regexp.MustCompile("b/[0-9a-z]+/b/[0-9a-z]+/t/([0-9a-z]+)")
	vhRegexp    = regexp.MustCompile("b/[0-9a-z]+/b/[0-9a-z]+/v/([0-9a-z]+)")
)

type StreamManager struct {
	consumer   transports.Consumer
	client     *rpc.Client
	backendURL string
	sub        types.Subscription
	ready      chan struct{}
	processed  uint64
	chainid    int
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
	client, err := rpc.Dial(backendURL)
	for err != nil {
		time.Sleep(1 * time.Second)
		client, err = rpc.Dial(backendURL)
	}
	if err != nil {
		return nil, err
	}
	var block miniBlock
	for i := 0; i < 720; i++ { // Retry for up to 1 hour (5 seconds * 720 = 3600 seconds = 1 hour, ignoring Call latency)
		err := client.Call(&block, "eth_getBlockByNumber", "latest", false)
		if err == nil {
			break
		}
		log.Warn("Failed to get initial block. Retrying.", "err", err, "retries", i)
		time.Sleep(5 * time.Second)
	}
	var chainid hexutil.Uint64
	if err := client.Call(&chainid, "eth_chainId"); err != nil {
		return nil, err
	}
	processed := uint64(0)
	lastNum := int64(block.Number)
	lastHash := types.Hash(block.Hash)
	lastWeight := new(big.Int).Add(block.Td.ToInt(), big.NewInt(int64(block.Number)))
	resumption, err := transports.ResumptionForTimestamp(brokerParams, int64(int(block.Timestamp)-rollbackInSeconds)*1000)
	if err != nil {
		log.Warn("Could not generate resumption token", "err", err)
	}

	var consumer transports.Consumer
	// It should be safe to calculate resumption weight this way. The producer
	// won't start publishing until the merge block, at which point it will set
	// the weight to Td + block Number. If we start up before the merge,
	consumer, err = transports.ResolveMuxConsumer(brokerParams, resumption, lastNum, lastHash, lastWeight, 128, trackedPrefixes, whitelist)
	if err != nil {
		return nil, err
	}
	return &StreamManager{
		consumer:   consumer,
		client:     client,
		ready:      make(chan struct{}, 1),
		chainid:    int(chainid),
		backendURL: backendURL,
		processed:  processed,
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
	ch := make(chan *delivery.ChainUpdate)
	m.sub = m.consumer.Subscribe(ch)
	go func() {
		<-m.consumer.Ready()
		m.ready <- struct{}{}
	}()
	go func() {
		for {
			log.Debug("Waiting for message")
			processed := uint64(0)
			select {
			case update := <-ch:
				// start := time.Now()
				forkChoiceMethod := "engine_forkchoiceUpdatedV1"
				added := update.Added()
				var latestHash, safeHash, finalHash common.Hash
				for _, pb := range added {
					var params beacon.ExecutableData
					if err := rlp.DecodeBytes(pb.Values[fmt.Sprintf("b/%x/b/%x/h", m.chainid, pb.Hash.Bytes())], &params); err != nil {
						log.Warn("Failed to rlp Decode block", "number", pb.Number, "err", err)
						continue
					}
					txs := make(map[int][]byte)
					versionedHashesMap := make(map[int]common.Hash)
					for k, v := range pb.Values {
						switch {
						case txRegexp.MatchString(k):
							parts := txRegexp.FindSubmatch([]byte(k))
							txIndex, _ := strconv.ParseInt(string(parts[1]), 16, 64)
							txs[int(txIndex)] = v
						case vhRegexp.MatchString(k):
							parts := vhRegexp.FindSubmatch([]byte(k))
							idx, _ := strconv.ParseInt(string(parts[1]), 16, 64)
							versionedHashesMap[int(idx)] = common.BytesToHash(v)
						}
					}
					versionedHashes := make([]common.Hash, len(versionedHashesMap))
					for i := 0; i < len(versionedHashesMap); i++ {
						versionedHashes[i] = versionedHashesMap[i]
					}
					params.Transactions = make([][]byte, len(txs))
					for i := 0; i < len(txs); i++ {
						params.Transactions[i] = txs[i]
					}
					ep := encodeParams(params)
					log.Info("Calling engine_newPayload", "block", pb.Number)
					var res interface{}
					if beaconRootBytes, ok := pb.Values[fmt.Sprintf("b/%x/b/%x/br", m.chainid, params.BlockHash.Bytes())]; ok {
						beaconRoot := common.BytesToHash(beaconRootBytes)
						forkChoiceMethod = "engine_forkchoiceUpdatedV3"
						log.Debug("Calling engine_newPayloadV3", "ep", ep, "vh", versionedHashes, "br", beaconRoot)
						if err := m.client.Call(&res, "engine_newPayloadV3", ep, versionedHashes, beaconRoot); err != nil {
							log.Warn("Error sending payload", "block", pb.Number, "err", err)
							errCh <- err
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
									return
								}
							} else {
								log.Warn("Error sending payload", "block", pb.Number, "err", err)
								errCh <- err
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
					HeadBlockHash:      latestHash,
					SafeBlockHash:      safeHash,
					FinalizedBlockHash: finalHash,
				}, nil); err != nil {
					log.Warn("Error sending forkchoice payload", "hash", latestHash, "err", err)
					errCh <- err
					return
				}
				if fcr.PayloadStatus.Status == beacon.VALID {
					m.processed += processed
				}
			}
		}
	}()
	if err := m.consumer.Start(); err != nil {
		errCh <- err
	}
	return errCh
}

func (m *StreamManager) Ready() chan struct{} {
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

type api struct {
	consumer transports.Consumer
}

func (a *api) WhyNotReady(hash types.Hash) string {
	return a.consumer.WhyNotReady(hash)
}
