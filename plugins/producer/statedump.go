package producer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	cli "github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/rlp"
)

var (
	snapRootKey = []byte("SnapshotRoot")
	snapshotAccountPrefix = []byte("a")
	snapshotStoragePrefix = []byte("o")
	codePrefix            = []byte("c")
)

type output struct{
	Key string `json:"key"`
	Value hexutil.Bytes `json:"value"`
}

type blockMetaOutput struct{
	Hash       common.Hash `json:"hash"`
	ParentHash common.Hash `json:"parentHash"`
	Number     uint64      `json:"number"`
	Weight     hexutil.Big `json:"weight"`
}

type acct struct {
	Nonce    uint64
	Balance  *big.Int
	Root     []byte
	CodeHash []byte
}

func (a *acct) slimRLP() ([]byte, error) {
	slim := &acct{
		Nonce: a.Nonce,
		Balance: a.Balance,
	}
	if !bytes.Equal(a.Root, emptyRoot[:]) {
		slim.Root = a.Root
	}
	if !bytes.Equal(a.CodeHash, emptyCode[:]) {
		slim.CodeHash = a.CodeHash
	}
	return rlp.EncodeToBytes(slim)
}

// fullAccount decodes the data on the 'slim RLP' format and return
// the consensus format account.
func fullAccount(data []byte) (acct, error) {
	var account acct
	if err := rlp.DecodeBytes(data, &account); err != nil {
		return acct{}, err
	}
	if len(account.Root) == 0 {
		account.Root = emptyRoot[:]
	}
	if len(account.CodeHash) == 0 {
		account.CodeHash = emptyCode[:]
	}
	return account, nil
}

var (
	// emptyRoot is the known root hash of an empty trie.
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421").Bytes()
	// emptyCode is the known hash of the empty EVM bytecode.
	emptyCode = crypto.Keccak256(nil)
	Subcommands = map[string]func(cli.Context, []string) error {
		"triedump": trieDump,
		"statedump": func(cli.Context, []string) error {
			log.Info("Starting state dump")
			db := backend.ChainDb()
			snaprootbytes, _ := db.Get(snapRootKey)
			snaproot := common.BytesToHash(snaprootbytes)
			header := backend.CurrentHeader()
			log.Info("Starting state dump", "headBlock", header.Number.Uint64(), "snaproot", snaproot)
			for header.Root != snaproot {
				var err error
				header, err = backend.HeaderByNumber(context.Background(), rpc.BlockNumber(header.Number.Int64() - 1))
				if err != nil { return err }
			}
			blockno := uint64(header.Number.Int64())
			td := backend.GetTd(context.Background(), header.Hash())

			acctIter := db.NewIterator(snapshotAccountPrefix, nil)
			defer acctIter.Release()
			jsonStream := json.NewEncoder(os.Stdout)
			jsonStream.Encode(blockMetaOutput{
				Hash: header.Hash(),
				ParentHash: header.ParentHash,
				Number: blockno,
				Weight: hexutil.Big(*td),
			})
			headerBytes, err := rlp.EncodeToBytes(header)
			if err != nil {
				log.Error("error encoding header to bytes, producer", "err", err)
				return err
			}
			jsonStream.Encode(output{Key: fmt.Sprintf("c/%x/b/%x/h", chainid, header.Hash().Bytes()), Value: headerBytes})
			log.Info("Dumping state for block", "num", blockno, "hash", header.Hash())
			for acctIter.Next() {
				if len(acctIter.Key()) != 33 { continue }
				hashedAddress := acctIter.Key()[1:]
				acctKey := fmt.Sprintf("c/%x/a/%x/d", chainid, hashedAddress)
				jsonStream.Encode(output{Key: acctKey, Value: hexutil.Bytes(acctIter.Value())})
				acct, err := fullAccount(acctIter.Value())
				if err != nil {
					log.Crit("Error decoding account", "acct", fmt.Sprintf("%#x", hashedAddress), "k", hexutil.Bytes(acctIter.Key()), "rlp", hexutil.Bytes(acctIter.Value()), "err", err)
					return err
				}
				if !bytes.Equal(acct.CodeHash, emptyCode) {
					v, err := db.Get(append(codePrefix, acct.CodeHash...))
					if len(v) == 0 {
						v, err = db.Get(acct.CodeHash)
					}
					if err != nil { return err }
					codeKey := fmt.Sprintf("c/%x/c/%x", chainid, acct.CodeHash)
					jsonStream.Encode(output{Key: codeKey, Value: hexutil.Bytes(v)})
				}
				if !bytes.Equal(acct.Root, emptyRoot) {
					count := 0
					slotIter := db.NewIterator(append(snapshotStoragePrefix, hashedAddress...), nil)
					for slotIter.Next() {
						count++
						if len(slotIter.Key()) != 65 { continue }
						slot := slotIter.Key()[33:]
						slotKey := fmt.Sprintf("c/%x/a/%x/s/%x", chainid, hashedAddress, slot)
						jsonStream.Encode(output{Key: slotKey, Value: hexutil.Bytes(slotIter.Value())})
					}
					if err := slotIter.Error(); err != nil { return err }
					slotIter.Release()
					if count == 0 { log.Warn("Found 0 slots for non-empty account")}
				}
			}
			if err := acctIter.Error(); err != nil { return err }
			return nil
		},
	}
)
