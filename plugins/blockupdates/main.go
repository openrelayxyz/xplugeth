package blockupdates

import (
	"fmt"
	// "strings"
	// "bytes"
	"io"
	"context"
	// "time"
	"encoding/json"
	"math/big"
	lru "github.com/hashicorp/golang-lru"
	

	"github.com/openrelayxyz/xplugeth"
	"github.com/openrelayxyz/xplugeth/types"
	
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
)


var (
	backend types.Backend
// 	lastBlock core.Hash
	cache *lru.Cache
	recentEmits *lru.Cache
// 	snapshotFlagName = "snapshot"
// 	log core.Logger
// 	blockEvents core.Feed
	suCh chan *stateUpdateWithRoot
)


// stateUpdate will be used to track state updates
type stateUpdate struct {
	Destructs map[common.Hash]struct{}
	Accounts map[common.Hash][]byte
	Storage map[common.Hash]map[common.Hash][]byte
	Code map[common.Hash][]byte
}

type stateUpdateWithRoot struct {
	su *stateUpdate
	root common.Hash
}

type BlockUpdates struct{
	backend types.Backend
}



// kvpair is used for RLP encoding of maps, as maps cannot be RLP encoded directly
type kvpair struct {
	Key common.Hash
	Value []byte
}

// storage is used for RLP encoding two layers of maps, as maps cannot be RLP encoded directly
type storage struct {
	Account common.Hash
	Data []kvpair
}

// storedStateUpdate is an RLP encodable version of stateUpdate
type storedStateUpdate struct {
	Destructs []common.Hash
	Accounts	[]kvpair
	Storage	 []storage
	Code	[]kvpair
}


// MarshalJSON represents the stateUpdate as JSON for RPC calls
func (su *stateUpdate) MarshalJSON() ([]byte, error) {
	result := make(map[string]interface{})
	destructs := make([]common.Hash, 0, len(su.Destructs))
	for k := range su.Destructs {
		destructs = append(destructs, k)
	}
	result["destructs"] = destructs
	accounts := make(map[string]hexutil.Bytes)
	for k, v := range su.Accounts {
		accounts[k.String()] = hexutil.Bytes(v)
	}
	result["accounts"] = accounts
	storage := make(map[string]map[string]hexutil.Bytes)
	for m, s := range su.Storage {
		storage[m.String()] = make(map[string]hexutil.Bytes)
		for k, v := range s {
			storage[m.String()][k.String()] = hexutil.Bytes(v)
		}
	}
	result["storage"] = storage
	code := make(map[string]hexutil.Bytes)
	for k, v := range su.Code {
		code[k.String()] = hexutil.Bytes(v)
	}
	result["code"] = code
	return json.Marshal(result)
}

// EncodeRLP converts the stateUpdate to a storedStateUpdate, and RLP encodes the result for storage
func (su *stateUpdate) EncodeRLP(w io.Writer) error {
	destructs := make([]common.Hash, 0, len(su.Destructs))
	for k := range su.Destructs {
		destructs = append(destructs, k)
	}
	accounts := make([]kvpair, 0, len(su.Accounts))
	for k, v := range su.Accounts {
		accounts = append(accounts, kvpair{k, v})
	}
	s := make([]storage, 0, len(su.Storage))
	for a, m := range su.Storage {
		accountStorage := storage{a, make([]kvpair, 0, len(m))}
		for k, v := range m {
			accountStorage.Data = append(accountStorage.Data, kvpair{k, v})
		}
		s = append(s, accountStorage)
	}
	code := make([]kvpair, 0, len(su.Code))
	for k, v := range su.Code {
		code = append(code, kvpair{k, v})
	}
	return rlp.Encode(w, storedStateUpdate{destructs, accounts, s, code})
}

// DecodeRLP takes a byte stream, decodes it to a storedStateUpdate, the n converts that into a stateUpdate object
func (su *stateUpdate) DecodeRLP(s *rlp.Stream) error {
	ssu := storedStateUpdate{}
	if err := s.Decode(&ssu); err != nil { return err }
	su.Destructs = make(map[common.Hash]struct{})
	for _, s := range ssu.Destructs {
		su.Destructs[s] = struct{}{}
	}
	su.Accounts = make(map[common.Hash][]byte)
	for _, kv := range ssu.Accounts {
		su.Accounts[kv.Key] = kv.Value
	}
	su.Storage = make(map[common.Hash]map[common.Hash][]byte)
	for _, s := range ssu.Storage {
		su.Storage[s.Account] = make(map[common.Hash][]byte)
		for _, kv := range s.Data {
			su.Storage[s.Account][kv.Key] = kv.Value
		}
	}
	su.Code = make(map[common.Hash][]byte)
	for _, kv := range ssu.Code {
		su.Code[kv.Key] = kv.Value
	}
	return nil
}

// var (
// 	httpApiFlagName = "http.api"
// 	wsApiFlagName = "ws.api"
// )

func init() {
	xplugeth.RegisterModule[BlockUpdates]()
}

// // Initialize does initial setup of variables as the plugin is loaded.
// func Initialize(ctx core.Context, loader core.PluginLoader, logger core.Logger) {
// 	log = logger
// 	pl = loader
// 	blockEvents = pl.GetFeed()
// 	cache, _ = lru.New(128)
// 	recentEmits, _ = lru.New(128)
// 	suCh = make(chan *stateUpdateWithRoot, 128)
// 	if !ctx.Bool(snapshotFlagName) {
// 		log.Warn("Snapshots are required for StateUpdate plugins, but are currently disabled. State Updates will be unavailable")
// 	}
// 	v := ctx.String(httpApiFlagName)
// 	if v == "" {
// 		ctx.Set(httpApiFlagName, "eth,net,web3,plugeth")
// 	} else if !strings.Contains(v, "plugeth") {
// 		ctx.Set(httpApiFlagName, v+",plugeth")
// 	}
// 	v = ctx.String(wsApiFlagName)
// 	if v == "" {
// 		ctx.Set(wsApiFlagName, "eth,net,web3,plugeth")
// 	} else if !strings.Contains(v, "plugeth") {
// 		ctx.Set(wsApiFlagName, v+",plugeth")
// 	}
// 	log.Info("Loaded block updater plugin")
// 	go func () {
// 		for su := range suCh {
// 			data, err := rlp.EncodeToBytes(su.su)
// 			if err != nil {
// 				log.Error("Failed to encode state update", "root", su.root, "err", err)
// 			}
// 			if err := backend.ChainDb().Put(append([]byte("su"), su.root.Bytes()...), data); err != nil {
// 				log.Error("Failed to store state update", "root", su.root, "err", err)
// 			}
// 			log.Debug("Stored state update", "root", su.root)
// 		}
// 	}()
// }


// InitializeNode is invoked by the plugin loader when the node and Backend are
// ready. We will track the backend to provide access to blocks and other
// useful information.
func (bu *BlockUpdates) InitializeNode(stack *node.Node, b types.Backend) {
	bu.backend = b
	backend = b
	cache, _ = lru.New(128)
	log.Error("Initialized node block updater plugin")
}


// // StateUpdate gives us updates about state changes made in each block. We
// // cache them for short term use, and write them to disk for the longer term.
func (*BlockUpdates) StateUpdate(blockRoot, parentRoot common.Hash, destructs map[common.Hash]struct{}, accounts map[common.Hash][]byte, storage map[common.Hash]map[common.Hash][]byte, codeUpdates map[common.Hash][]byte) {
	if backend == nil {
		log.Warn("State update called before InitializeNode", "root", blockRoot)
		return
	}
	su := &stateUpdate{
		Destructs: destructs,
		Accounts: accounts,
		Storage: storage,
		Code: codeUpdates,
	}
	cache.Add(blockRoot, su)
	suCh <- &stateUpdateWithRoot{su: su, root: blockRoot}
}

// // AppendAncient removes our state update records from leveldb as the
// // corresponding blocks are moved from leveldb to the ancients database. At
// // some point in the future, we may want to look at a way to move the state
// // updates to an ancients table of their own for longer term retention.
// func AppendAncient(number uint64, hash, headerBytes, body, receipts, td []byte) {
// 	header := new(types.Header)
// 	if err := rlp.Decode(bytes.NewReader(headerBytes), header); err != nil {
// 		log.Warn("Could not decode ancient header", "block", number)
// 		return
// 	}
// 	go func() {
// 		// Background this so we can clean up once the backend is set, but we don't
// 		// block the creation of the backend.
// 		for backend == nil {
// 			time.Sleep(250 * time.Millisecond)
// 		}
// 		backend.ChainDb().Delete(append([]byte("su"), header.Root.Bytes()...))
// 	}()

// }

// NewHead is invoked when a new block becomes the latest recognized block. We
// use this to notify the blockEvents channel of new blocks, as well as invoke
// the BlockUpdates hook on downstream plugins.
// TODO: We're not necessarily handling reorgs properly, which may result in
// some blocks not being emitted through this hook.
func (*BlockUpdates) NewHead(block *gtypes.Block, hash common.Hash, logs []*gtypes.Log, td *big.Int) {
	log.Error("INSIDE OF NEWW HEAD")
	// newHead(block, hash, td)
}
// func newHead(block types.Block, hash core.Hash, td *big.Int) {
// 	if recentEmits.Contains(hash) {
// 		log.Debug("Skipping recently emitted block")
// 		return
// 	}
// 	result, err := blockUpdates(context.Background(), &block)
// 	if err != nil {
// 		log.Error("Could not serialize block", "err", err, "hash", block.Hash())
// 		return
// 	}
// 	if recentEmits.Len() > 10 && !recentEmits.Contains(block.ParentHash()) {
// 		blockRLP, err := backend.BlockByHash(context.Background(), block.ParentHash())
// 		if err != nil {
// 			log.Error("Could not get block for reorg", "hash", block.ParentHash(), "err", err)
// 			return
// 		}
// 		var parentBlock types.Block
// 		if err := rlp.DecodeBytes(blockRLP, &parentBlock); err != nil {
// 			log.Error("Could not decode block during reorg", "hash", block.ParentHash(), "err", err)
// 			return
// 		}
// 		td := backend.GetTd(context.Background(), parentBlock.Hash())
// 		newHead(parentBlock, block.Hash(), td)
// 	}
// 	blockEvents.Send(result)

// 	receipts := result["receipts"].(types.Receipts)
// 	su := result["stateUpdates"].(*stateUpdate)
// 	fnList := pl.Lookup("BlockUpdates", func(item interface{}) bool {
// 		_, ok := item.(func(*types.Block, *big.Int, types.Receipts, map[core.Hash]struct{}, map[core.Hash][]byte, map[core.Hash]map[core.Hash][]byte, map[core.Hash][]byte))
// 		log.Info("Found BlockUpdates hook", "matches", ok)
// 		return ok
// 	})
// 	for _, fni := range fnList {
// 		if fn, ok := fni.(func(*types.Block, *big.Int, types.Receipts, map[core.Hash]struct{}, map[core.Hash][]byte, map[core.Hash]map[core.Hash][]byte, map[core.Hash][]byte)); ok {
// 			fn(&block, td, receipts, su.Destructs, su.Accounts, su.Storage, su.Code)
// 		}
// 	}
// 	recentEmits.Add(hash, struct{}{})
// }
func (*BlockUpdates) Reorg(common common.Hash, oldChain []common.Hash, newChain []common.Hash) {
	// fnList := pl.Lookup("BUPreReorg", func(item interface{}) bool {
	// 	_, ok := item.(func(core.Hash, []core.Hash, []core.Hash))
	// 	return ok
	// })
	// for _, fni := range fnList {
	// 	if fn, ok := fni.(func(core.Hash, []core.Hash, []core.Hash)); ok {
	// 		fn(common, oldChain, newChain)
	// 	}
	// }
	// for i := len(newChain) - 1; i >= 0; i-- {
	// 	blockHash := newChain[i]
	// 	blockRLP, err := backend.BlockByHash(context.Background(), blockHash)
	// 	if err != nil {
	// 		log.Error("Could not get block for reorg", "hash", blockHash, "err", err)
	// 		return
	// 	}
	// 	var block types.Block
	// 	if err := rlp.DecodeBytes(blockRLP, &block); err != nil {
	// 		log.Error("Could not decode block during reorg", "hash", blockHash, "err", err)
	// 		return
	// 	}
	// 	td := backend.GetTd(context.Background(), blockHash)
	// 	newHead(block, blockHash, td)
	// }
	// fnList = pl.Lookup("BUPostReorg", func(item interface{}) bool {
	// 	_, ok := item.(func(core.Hash, []core.Hash, []core.Hash))
	// 	return ok
	// })
	// for _, fni := range fnList {
	// 	if fn, ok := fni.(func(core.Hash, []core.Hash, []core.Hash)); ok {
	// 		fn(common, oldChain, newChain)
	// 	}
	// }
}


// BlockUpdates is a service that lets clients query for block updates for a
// given block by hash or number, or subscribe to new block upates.

// func BlockUpdatesByNumber(number int64) (*types.Block, *big.Int, types.Receipts, map[core.Hash]struct{}, map[core.Hash][]byte, map[core.Hash]map[core.Hash][]byte, map[core.Hash][]byte, error) {
// 	blockBytes, err := backend.BlockByNumber(context.Background(), int64(number))
// 	if err != nil { return nil, nil, nil, nil, nil, nil, nil, err }
// 	var block types.Block
// 	if err := rlp.DecodeBytes(blockBytes, &block); err != nil {
// 		return nil, nil, nil, nil, nil, nil, nil, err
// 	}
// 	td := backend.GetTd(context.Background(), block.Hash())
// 	receiptBytes, err := backend.GetReceipts(context.Background(), block.Hash())
// 	if err != nil { return nil, nil, nil, nil, nil, nil, nil, err }
// 	var receipts types.Receipts
// 	if err := json.Unmarshal(receiptBytes, &receipts); err != nil {
// 		return nil, nil, nil, nil, nil, nil, nil, err
// 	}
// 	var su *stateUpdate
// 	if v, ok := cache.Get(block.Root()); ok {
// 		su = v.(*stateUpdate)
// 	} else {
// 		su = new(stateUpdate)
// 		data, err := backend.ChainDb().Get(append([]byte("su"), block.Root().Bytes()...))
// 		if err != nil { return &block, td, receipts, nil, nil, nil, nil, fmt.Errorf("State Updates unavailable for block %v", block.Hash())}
// 		if err := rlp.DecodeBytes(data, su); err != nil { return &block, td, receipts, nil, nil, nil, nil, fmt.Errorf("State updates unavailable for block %#x", block.Hash()) }
// 	}
// 	return &block, td, receipts, su.Destructs, su.Accounts, su.Storage, su.Code, nil
// }

// blockUpdate handles the serialization of a block
func blockUpdates(ctx context.Context, block *gtypes.Block) (map[string]interface{}, error)	{
	result, err := RPCMarshalBlock(block, true, true)
	if err != nil { return nil, err }
	result["receipts"], err = backend.GetReceipts(ctx, block.Hash())
	if err != nil { return nil, err }
	if v, ok := cache.Get(block.Root()); ok {
		result["stateUpdates"] = v
		return result, nil
	}
	data, err := backend.ChainDb().Get(append([]byte("su"), block.Root().Bytes()...))
	if err != nil { return nil, fmt.Errorf("State Updates unavailable for block %#x", block.Hash())}
	log.Error("its the second one")
	su := &stateUpdate{}
	if err := rlp.DecodeBytes(data, su); err != nil { return nil, fmt.Errorf("State updates unavailable for block %#x", block.Hash()) }
	result["stateUpdates"] = su
	return result, nil
}

// BlockUpdatesByNumber retrieves a block by number, gets receipts and state
// updates, and serializes the response.
func (b *BlockUpdates) BlockUpdatesByNumber(ctx context.Context, number rpc.BlockNumber) (map[string]interface{}, error) {
	block, err := b.backend.BlockByNumber(ctx, number)
	if err != nil { return nil, err }
	return blockUpdates(ctx, block)
}

// BlockUpdatesByHash retrieves a block by hash, gets receipts and state
// updates, and serializes the response.
func (b *BlockUpdates) BlockUpdatesByHash(ctx context.Context, hash common.Hash) (map[string]interface{}, error) {
	block, err := b.backend.BlockByHash(ctx, hash)
	if err != nil { return nil, err }
	return blockUpdates(ctx, block)
}


// // BlockUpdates allows clients to subscribe to notifications of new blocks
// // along with receipts and state updates.
// func (b *BlockUpdates) BlockUpdates(ctx context.Context) (<-chan map[string]interface{}, error) {
// 	blockDataChan := make(chan map[string]interface{}, 1000)
// 	ch := make(chan map[string]interface{}, 1000)
// 	sub := blockEvents.Subscribe(blockDataChan)
// 	go func() {
// 		log.Info("BlockUpdates subscription setup")
// 		defer log.Info("BlockUpdates subscription closed")
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				sub.Unsubscribe()
// 				close(ch)
// 				close(blockDataChan)
// 				return
// 			case b := <-blockDataChan:
// 				ch <- b
// 			}
// 		}
// 	}()
// 	return ch, nil
// }


// GetAPIs exposes the BlockUpdates service under the cardinal namespace.
func (*BlockUpdates) GetAPIs(stack *node.Node, backend types.Backend) []rpc.API {
	log.Error("GETAPIS", "backend", backend)
	return []rpc.API{
	 {
		 Namespace: "plugeth",
		 Version:	 "1.0",
		 Service:	 &BlockUpdates{backend},
		 Public:		true,
	 },
 }
}

func (bu *BlockUpdates) Blockchain() {
	be := bu.backend 
	log.Error("inside of blockchain", "backend", be)
}

func (bu *BlockUpdates) TestBU(context.Context) string {
	var b bool
	if bu.backend != nil {
		b = true
	}
	return fmt.Sprintf("BlockUpdates module value: backend is not nil = %v", b)
}
