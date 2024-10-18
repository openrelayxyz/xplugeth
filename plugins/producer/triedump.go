package producer

import (
	"fmt"
	"bytes"
	"context"
	"strconv"
	"github.com/hashicorp/golang-lru"
	cli "github.com/urfave/cli/v2"

	"github.com/ethereum/go-ethereum/common"
	gtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
)

var (
	headerCache *lru.Cache
)

func getTrie(hash common.Hash) (state.Trie, error) {
	stateDB, _, err := backend.StateAndHeaderByNumberOrHash(context.Background(), rpc.BlockNumberOrHashWithHash(hash, true))
	if err != nil {
		return nil ,err
	}
	return stateDB.GetTrie(), nil
}

func traverseIterators(a, b trie.NodeIterator, add, delete func(k, v []byte), alter func(k1, v1, k2, v2 []byte)) {
	// Advance both iterators initially
	hasA := a.Next(true)
	hasB := b.Next(true)
	counter := 0

	for hasA && hasB {
		counter++
		switch compareNodes(a, b) {
		case -1: // a is behind b
			if a.Leaf() {
				delete(a.LeafKey(), a.LeafBlob())
			}
			hasA = a.Next(true) // advance only a
		case 1: // a is ahead of b
			if b.Leaf() {
				add(b.LeafKey(), b.LeafBlob())
			}
			hasB = b.Next(true) // advance only d
		case 0: // nodes are equal
			if a.Leaf() && b.Leaf() {
				if !bytes.Equal(a.LeafBlob(), b.LeafBlob()) {
					alter(a.LeafKey(), a.LeafBlob(), b.LeafKey(), b.LeafBlob())
					// add(b.LeafKey(), b.LeafBlob())
				}
			}
			descend := a.Hash() != b.Hash()
			if !descend && a.Hash() == (common.Hash{}) {
				descend = true
			}
			hasA = a.Next(descend)
			hasB = b.Next(descend) // advance both iterators
		}
	}

	// Handle remaining nodes in A
	for hasA {
		counter++
		if a.Leaf() {
			delete(a.LeafKey(), a.LeafBlob())
		}
		hasA = a.Next(true)
	}

	// Handle remaining nodes in B
	for hasB {
		counter++
		if b.Leaf() {
			add(b.LeafKey(), b.LeafBlob())
		}
		hasB = b.Next(true)
	}
	log.Info("Processed tries.", "nodes", counter)
}

func stateTrieUpdatesByNumber(i int64) (map[common.Hash]struct{}, map[common.Hash][]byte, map[common.Hash]map[common.Hash][]byte, map[common.Hash][]byte, error) {
	if headerCache == nil {
		headerCache, _ = lru.New(64)
	}
	destructs := make(map[common.Hash]struct{})
	accounts := make(map[common.Hash][]byte)
	storage := make(map[common.Hash]map[common.Hash][]byte)
	code := make(map[common.Hash][]byte)
	db := backend.ChainDb()

	var lastHeader, header *gtypes.Header
	var err error

	if v, ok := headerCache.Get(i-1); ok {
		lastHeader = v.(*gtypes.Header)
	} else {
		lastHeader, err := backend.HeaderByNumber(context.Background(), rpc.BlockNumber(i-1))
		if err != nil {
			log.Warn("Error getting starting header")
			return nil, nil, nil, nil, err
		}
		headerCache.Add(i-1, lastHeader)
	}
	var lastTrie state.Trie
	if lastHeader != nil {
		lastTrie, err = getTrie(lastHeader.Root)
	} else {
		return nil, nil, nil, nil, err
	}
	if err != nil {
		log.Error("Error getting trie", "block", startBlock)
		return nil, nil, nil, nil, err
	}

	if v, ok := headerCache.Get(i); ok {
		header = v.(*gtypes.Header)
	} else {
		header = &gtypes.Header{}
		header, err := backend.HeaderByNumber(context.Background(), rpc.BlockNumber(i))
		if err != nil {
			log.Error(fmt.Sprintf("Error acquiring header for block %v", i), "err", err)
			return nil, nil, nil, nil, err
		}
		headerCache.Add(i, header)
	}
	var currentTrie state.Trie
	if header != nil {
		currentTrie, err = getTrie(header.Root)
	} else {
		return nil, nil, nil, nil, err
	}
	if err != nil {
		log.Error("Error getting last trie")
		return nil, nil, nil, nil, err
	}
	a, err := lastTrie.NodeIterator(nil)
	if err != nil {
		log.Error("Error getting node iterator from last trie, stateTrieUpdatesByNumber, producer")
		return nil, nil, nil, nil, err
	}
	b, err := currentTrie.NodeIterator(nil)
	if err != nil {
		log.Error("Error getting node iterator from last trie, stateTrieUpdatesByNumber, producer")
		return nil, nil, nil, nil, err
	}
	alteredAccounts := map[string]acct{}
	oldAccounts := map[string]acct{}
	codeChanges := map[common.Hash]struct{}{}


	traverseIterators(a, b, func(k, v []byte){
		// Added accounts
		account, err := fullAccount(v)
		if err != nil {
			log.Warn("Found invalid account in acount trie")
			return
		}
		alteredAccounts[string(k)] = account
	}, 
	func(k, v []byte) {
		// Deleted accounts
		account, err := fullAccount(v)
		if err != nil {
			log.Warn("Found invalid account in acount trie")
			return
		}
		oldAccounts[string(k)] = account
	},
	func(oldk, oldv, newk, newv []byte) {
		account, err := fullAccount(oldv)
		if err != nil {
			log.Warn("Found invalid account in acount trie")
			return
		}
		oldAccounts[string(oldk)] = account
		account, err = fullAccount(newv)
		if err != nil {
			log.Warn("Found invalid account in acount trie")
			return
		}
		alteredAccounts[string(newk)] = account
	})
	storageChanges := map[string]map[string][]byte{}
	// TODO: Iteration of the altered accounts could potentially be parallelized
	for k, acct := range alteredAccounts {
		var oldStorageTrie state.Trie
		if oldAcct, ok := oldAccounts[k]; ok {
			delete(oldAccounts, k)
			if !bytes.Equal(oldAcct.CodeHash, acct.CodeHash) {
				codeChanges[common.BytesToHash(acct.CodeHash)] = struct{}{}
			}
			if bytes.Equal(acct.Root, oldAcct.Root) {
				// Storage didn't change
				continue
			}
			oldStorageTrie, err = getTrie(common.BytesToHash(oldAcct.Root))
		} else {
			if !bytes.Equal(acct.CodeHash, emptyCode) {
				codeChanges[common.BytesToHash(acct.CodeHash)] = struct{}{}
			}
			if bytes.Equal(acct.Root, emptyRoot) {
				// Storage for new account is empty
				continue
			}
			oldStorageTrie, err = getTrie(common.BytesToHash(emptyRoot))
		}
		storageTrie, err := getTrie(common.BytesToHash(acct.Root))
		if err != nil {
			log.Error("error getting trie for account", "acct", k)
			return nil, nil, nil, nil, err
		}
		c, err := oldStorageTrie.NodeIterator(nil)
		if err != nil {
			log.Error("Error getting node iterator from old storage trie, stateTrieUpdatesByNumber, producer")
			return nil, nil, nil, nil, err
		}
		d, err := storageTrie.NodeIterator(nil)
		if err != nil {
			log.Error("Error getting node iterator from storage trie, stateTrieUpdatesByNumber, producer")
			return nil, nil, nil, nil, err
		}
		storageChanges[k] = map[string][]byte{}
		traverseIterators(c, d, func(key, v []byte) {
			// Add Storage
			storageChanges[k][string(key)] = v

		}, func(key, v []byte) {
			// Delete Storage
			if _, ok := storageChanges[k][string(key)]; !ok {
				storageChanges[k][string(key)] = []byte{}
			}
		}, func(oldk, oldv, newk, newv []byte) {
			storageChanges[k][string(newk)] = newv
		})
	}
	for k := range oldAccounts {
		destructs[common.BytesToHash([]byte(k))] = struct{}{}
	}
	for k, v := range alteredAccounts {
		accounts[common.BytesToHash([]byte(k))], err = v.slimRLP()
		if err != nil {
			return nil, nil, nil, nil, err
		}
		storage[common.BytesToHash([]byte(k))] = map[common.Hash][]byte{}
		for sk, sv := range storageChanges[k] {
			storage[common.BytesToHash([]byte(k))][common.BytesToHash([]byte(sk))] = sv
		}
	}
	for codeHash := range codeChanges {
		if !bytes.Equal(codeHash.Bytes(), emptyCode) {
			c, _ := db.Get(append(codePrefix, codeHash.Bytes()...))
			if len(c) == 0 {
				c, err = db.Get(codeHash.Bytes())
				if err != nil {
					return nil, nil, nil, nil, err
				}
			}
			code[common.BytesToHash(codeHash.Bytes())] = c
		}
	}
	return destructs, accounts, storage, code, nil
}


func trieDump (ctx cli.Context, args []string) error {
	log.Info("Starting trie dump")
	// chainConfig := backend.ChainConfig()
	header := backend.CurrentHeader()
	startBlock := int64(0)
	endBlock := header.Number.Int64()
	if len(args) > 0 {
		s, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		startBlock = int64(s)
	}
	if len(args) > 1 {
		e, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		endBlock = int64(e)
	}
	lastHeader, err := backend.HeaderByNumber(context.Background(), rpc.BlockNumber(startBlock))
	if err != nil {
		log.Warn("Error getting starting header")
		return err
	}
	var lastTrie state.Trie
	lastTrie, err = getTrie(lastHeader.Root)
	if err != nil {
		log.Error("error returned acquiring trie for last header root stateTrieUpdatesByNumber, producer", "err", err)
	}
	for i := startBlock+1; i <= endBlock; i++ {
		header, err := backend.HeaderByNumber(context.Background(), rpc.BlockNumber(i))
		if err != nil {
			log.Error(fmt.Sprintf("error returned getting header for block number %v stateTrieUpdatesByNumber, producer", i), "err", err)
			return err
		}
		var currentTrie state.Trie
		currentTrie, err = getTrie(header.Root)
		if err != nil {
			log.Error("Error getting last trie")
			return err
		}

		a, err := lastTrie.NodeIterator(nil)
		if err != nil {
			log.Error("Error getting node iterator from last trie")
			return err
		}
		b, err := currentTrie.NodeIterator(nil)
		if err != nil {
			log.Error("Error getting node iterator from current trie")
			return err
		}

		alteredAccounts := map[string]acct{}
		oldAccounts := map[string]acct{}


		COMPARE_NODES:
		for {
			switch compareNodes(a, b) {
			case -1:
				// Node exists in lastTrie but not currentTrie
				// This is a deletion
				if a.Leaf() {
					account, err := fullAccount(a.LeafBlob())
					if err != nil {
						log.Warn("Found invalid account in acount trie")
						continue
					}
					oldAccounts[string(a.LeafKey())] = account
				}

				// b jumped past a; advance a
				a.Next(true)
			case 1:
				// Node exists in currentTrie but not lastTrie
				// This is an addition

				if b.Leaf() {
					account, err := fullAccount(b.LeafBlob())
					if err != nil {
						log.Warn("Found invalid account in acount trie")
						continue
					}
					alteredAccounts[string(b.LeafKey())] = account
				}

				if !b.Next(true) {
					break COMPARE_NODES
				}

			case 0:
				// a and b are identical; skip this whole subtree if the nodes have hashes
				hasHash := a.Hash() == common.Hash{}
				if !b.Next(hasHash) {
					break COMPARE_NODES
				}
				if !a.Next(hasHash) {
					break COMPARE_NODES
				}
			}
		}
		storageChanges := map[string]map[string][]byte{}
		for k, acct := range alteredAccounts {
			var oldStorageTrie state.Trie
			if oldAcct, ok := oldAccounts[k]; ok {
				delete(oldAccounts, k)
				if bytes.Equal(acct.Root, oldAcct.Root) {
					// Storage didn't change
					continue
				}
				oldStorageTrie, err = getTrie(common.BytesToHash(oldAcct.Root))
				} else {
				oldStorageTrie, err = getTrie(common.BytesToHash(emptyRoot))
			}
			if bytes.Equal(acct.Root, emptyRoot) {
				// Empty trie, nothing to see here
				continue
			}
			storageTrie, err := getTrie(common.BytesToHash(acct.Root))
			if err != nil {
				log.Error("error getting trie for account", "acct", k)
				return err
			}
			c, err := oldStorageTrie.NodeIterator(nil)
			if err != nil {
				log.Error("Error getting node iterator from old storage trie")
				return err
			}
			d, err := storageTrie.NodeIterator(nil)
			if err != nil {
				log.Error("Error getting node iterator from storage trie")
				return err
			}
			storageChanges[k] = map[string][]byte{}
			COMPARE_STORAGE:
			for {
				switch compareNodes(c, d) {
				case -1:
					// Node exists in lastTrie but not currentTrie
					// This is a deletion
					if c.Leaf() {
						storageChanges[k][string(c.LeafKey())] = []byte{}
						// storageChanges[fmt.Sprintf("c/%x/c/%x", chainConfig.ChainID, []byte(k), c.LeafKey())] = []byte{}
					}

					// c jumped past d; advance d
					c.Next(true)
				case 1:
					// Node exists in currentTrie but not lastTrie
					// This is an addition

					if d.Leaf() {
						storageChanges[k][string(d.LeafKey())] = d.LeafBlob()
					}

					if !d.Next(true) {
						break COMPARE_STORAGE
					}

				case 0:
					// a and b are identical; skip this whole subtree if the nodes have hashes
					hasHash := c.Hash() == common.Hash{}
					if !d.Next(hasHash) {
						break COMPARE_STORAGE
					}
					if !c.Next(hasHash) {
						break COMPARE_STORAGE
					}
				}
			}

		}
		for k := range oldAccounts {
			log.Info("Destructed", "account", []byte(k))
		}
		for k, v := range alteredAccounts {
			log.Info("Altered", "block", i, "account", []byte(k), "data", v)
			for sk, sv := range storageChanges[k] {
				log.Info("Storage key", "key", sk, "val", sv)
			}
		}
		lastTrie = currentTrie
	}
	return nil
}


func compareNodes(a, b trie.NodeIterator) int {
	if cmp := bytes.Compare(a.Path(), b.Path()); cmp != 0 {
		return cmp
	}
	if a.Leaf() && !b.Leaf() {
		return -1
	} else if b.Leaf() && !a.Leaf() {
		return 1
	}
	// if cmp := bytes.Compare(a.Hash().Bytes(), b.Hash().Bytes()); cmp != 0 {
	// 	return cmp
	// }
	// if a.Leaf() && b.Leaf() {
	// 	return bytes.Compare(a.LeafBlob(), b.LeafBlob())
	// }
	return 0
}