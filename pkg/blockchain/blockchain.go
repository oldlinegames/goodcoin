package blockchain

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/block/tx/txi"
	"BrunoCoin/pkg/block/tx/txo"
	"BrunoCoin/pkg/proto"
	"fmt"
	"strings"
	"sync"
)

// BlockchainNode represents a collection of information
// relevant to one block in the chain.
// Block is the particular block
// PrevNode is the node that this block references before
// it
// utxo is a map of txo identifiers to transaction outputs
// It represents all UTXO on the chain of this block up
// until this block.
// depth is how far the block is down in its chain.
type BlockchainNode struct {
	*block.Block
	PrevNode *BlockchainNode
	utxo     map[string]*txo.TransactionOutput
	depth    int
}

// Blockchain not only stores the main blockchain, but
// it actually stores all forked blockchains in a tree
// like structure using a map.
// Addr is the address of the node storing the blockchain.
// blocks are all blocks (forked or not) stored in a tree
// using a map
// LastBlock is the last block of the main chain
type Blockchain struct {
	Addr      string
	blocks    map[string]*BlockchainNode
	LastBlock *BlockchainNode
	sync.Mutex
}

// New creates the initial blockchain with 1 starting block,
// which is the GENESIS_BLOCK. This block is static is
// hardcoded into every blockchain as the first block.
// Inputs:
// conf *Config the configuration for the blockchain.
func New(conf *Config) *Blockchain {
	genBlock := GenesisBlock(conf)
	genTx := genBlock.Transactions[0]
	genTxKey := fmt.Sprintf("%v-%v", genTx.Hash(), 0)
	GenesisBlock := &BlockchainNode{
		Block:    genBlock,
		PrevNode: nil,
		utxo:     map[string]*txo.TransactionOutput{genTxKey: genTx.Outputs[0]},
		depth:    0,
	}
	return &Blockchain{
		blocks:    map[string]*BlockchainNode{GenesisBlock.Hash(): GenesisBlock},
		LastBlock: GenesisBlock,
	}
}

// SetAddr sets the address of the node storing the
// blockchain as a field of the blockchain struct.
// Inputs:
// a string the address to be set
func (bc *Blockchain) SetAddr(a string) {
	bc.Lock()
	bc.Addr = a
	bc.Unlock()
}

// Add adds a block to the blockchain in the correct
// spot.
// Inputs:
// b *block.Block the block to be added
// TODO
// 1. Find previous node (that this block is being appended
// to)
// 2. From the previous node's utxo, remove any used utxo
// and add the new utxo from the new block
// 3. Craft a new blockchain node and put it in the correct
//// spot
// Tip 1: Remember that this function mutates state
// concurrently with other go routines
// Tip 2: It might be helpful to add a debug message
// after the block is successfully added

// some functions/fields/methods that might be helpful
// let b be a bloc object
// bc.Lock()
// bc.Unlock()
// utils.FmtAddr(...)
// b.NameTag()
// txo.MkTXOLoc(...)
func (bc *Blockchain) Add(b *block.Block) {
	if b == nil {
		return
	}
	bc.Lock()
	defer bc.Unlock()
	// 1. Find previous node (that this block is being appended
	// to)
	lastNode, ok := bc.blocks[b.Hdr.PrvBlkHsh]
	if !ok {
		// blocks prev node is not in the chain
		return
	}

	// 2. From the previous node's utxo, remove any used utxo
	// and add the new utxo from the new block

	utxoCopy := lastNode.utxo

	for _, transaction := range b.Transactions {
		// remove UTXO
		for _, txInput := range transaction.Inputs {
			key := txo.MkTXOLoc(txInput.TransactionHash, txInput.OutputIndex)
			delete(utxoCopy, key)
		}
		for ind, txOutput := range transaction.Outputs {
			key := txo.MkTXOLoc(transaction.Hash(), uint32(ind))
			utxoCopy[key] = txOutput
		}
	}
	// 3. Craft a new blockchain node and put it in the correct
	// spot
	bn := &BlockchainNode{
		b,
		lastNode,
		utxoCopy,
		lastNode.depth + 1,
	}
	bc.blocks[b.Hash()] = bn
	if bn.depth > bc.LastBlock.depth {
		bc.LastBlock = bn
	}

	return
}

// Length returns the count of blocks on the
// blockchain.
// Returns:
// int the number of blocks on the main chain.
func (bc *Blockchain) Length() int {
	bc.Lock()
	defer bc.Unlock()
	return bc.LastBlock.depth + 1
}

// Get returns the blocks that corresponds to a
// particular inputted hash
// Inputs:
// hash string the hash of the block wanting to
// be returned
// Returns:
// *block.Block the block corresponding to the hash
func (bc *Blockchain) Get(hash string) *block.Block {
	bc.Lock()
	defer bc.Unlock()
	return bc.blocks[hash].Block
}

// IndexOf (GetIndex) gets the index in the blockchain
// for a particular block (the hash of that block).
// Inputs:
// h string the hash of the block whose index is being
// searched for
// Returns:
// int the index of the block
func (bc *Blockchain) IndexOf(hash string) int {
	bc.Lock()
	defer bc.Unlock()
	if bc.blocks[hash] == nil {
		return -1
	}
	return bc.blocks[hash].depth
}

// GetLastBlock is a getter for LastBlock
// Returns:
// *block.Block the last block of the main chain.
func (bc *Blockchain) GetLastBlock() *block.Block {
	bc.Lock()
	defer bc.Unlock()
	return bc.LastBlock.Block
}

// List returns all blocks on the main chain in order.
// Returns:
// []*block.Block list of all blocks on main chain in order.
func (bc *Blockchain) List() []*block.Block {
	bc.Lock()
	defer bc.Unlock()
	b := bc.LastBlock
	slice := make([]*block.Block, 0)
	for ct := bc.LastBlock.depth + 1; ct > 0; ct-- {
		slice = append([]*block.Block{b.Block}, slice...)
		b = b.PrevNode
	}
	return slice
}

// Slice returns a slice of the main chain from a certain
// starting index to an ending index (exclusive).
// Inputs:
// s int the starting index
// e int the ending index (exclusive)
// Returns:
// []*block.Block the list of blocks on the main chain
// in order from starting index to ending index (exclusive)
func (bc *Blockchain) Slice(s int, e int) []*block.Block {
	bc.Lock()
	defer bc.Unlock()
	b := bc.LastBlock
	slice := make([]*block.Block, 0)
	for b.depth >= s {
		if b.depth < e {
			slice = append([]*block.Block{b.Block}, slice...)
		}
		if b.PrevNode == nil {
			break
		}
		b = b.PrevNode
	}
	return slice
}

// IsEndMainChain checks whether a new block would
// be appended to the end of the current chain.
// Inputs:
// blk *block.Block the new block that is going to be
// added to the main chain.
// Returns:
// bool True if the block would be appended to the main
// chain. False otherwise
func (bc *Blockchain) IsEndMainChain(blk *block.Block) bool {
	return bc.LastBlock.Block.Hash() == blk.Hdr.PrvBlkHsh
}

func (bc *Blockchain) GetUTXO(txi *txi.TransactionInput) *txo.TransactionOutput {
	bc.Lock()
	defer bc.Unlock()
	key := txo.MkTXOLoc(txi.TransactionHash, txi.OutputIndex)
	utxo, _ := bc.LastBlock.utxo[key]
	return utxo
}

func (bc *Blockchain) GetUTXOLen(pk string) int {
	bc.Lock()
	defer bc.Unlock()
	ct := 0
	for _, v := range bc.LastBlock.utxo {
		if v.LockingScript == pk {
			ct++
		}
	}
	return ct
}

// IsInvalidInput checks whether a transaction input
// is an orphan (whether its utxo exists).
// Inputs:
// txi *txi.TransactionInput the transaction input
// being tested for orphan-ness
// Returns:
// bool True if the transaction input is an orphan, false
// otherwise
func (bc *Blockchain) IsInvalidInput(txi *txi.TransactionInput) bool {
	bc.Lock()
	defer bc.Unlock()
	key := txo.MkTXOLoc(txi.TransactionHash, txi.OutputIndex)
	_, found := bc.LastBlock.utxo[key]
	return !found
}

// ChkChainsUTXO (checkchainsutxo) checks to see that
// the transactions all reference valid UTXO on whatever
// forked chain that the transactions belonging to a block
// are being added to.
// Inputs:
// txs []*tx.Transaction the txs on a new block wanting to
// be added to the chain.
// prevHash string the hash of the previous block that the
// the block with the inputted txs reference
// Returns:
// bool True if each input from the txs reference a valid
// utxo
func (bc *Blockchain) ChkChainsUTXO(txs []*tx.Transaction, prevHash string) bool {
	var keys []string
	lastBlock, found := bc.blocks[prevHash]
	// If not found, this is an orphan, try to compare it against the last node
	if !found {
		lastBlock = bc.LastBlock
	}
	for _, t := range txs {
		for _, txii := range t.Inputs {
			key := txo.MkTXOLoc(txii.TransactionHash, txii.OutputIndex)
			if _, found := lastBlock.utxo[key]; !found {
				return false
			}
			keys = append(keys, key)
		}
	}
	return true
}

// UTXOInfo holds the information about a utxo
// necessary for making a transaction input.
// TxHsh	the hash of the transaction that the utxo
// is from
// OutIdx	the index into the outputs array of the
// transaction that the utxo is from
// UTXO		the actual utxo object
// Amt		the amount of money in the utxo.
type UTXOInfo struct {
	TxHsh  string
	OutIdx uint32
	UTXO   *txo.TransactionOutput
	Amt    uint32
}

// GetUTXOForAmt (GetUTXOForAmount) gets
// enough utxo (if there is enough) for the
// inputted amount. This is used by the wallet
// to ask for enough utxo to make a transaction
// for a certain person.
// Inputs:
// amt uint32 the amount of money needed
// pubKey string the person that the utxo belongs
// to
// Returns:
// []*UTXOInfo the list of utxo information that
// is needed to construct transaction inputs for
// a transaction with the inputted amount
// uint32 the amount of change left over
// bool True if there is enough utxo for the amt,
// false otherwise.
// TODO
// 1. Find the utxo on the last block of the main chain.
// 2. For the utxo payable to the passed in pubkey, try
// and get enough for the inputted amount
// Tip 1: Remember that this method accesses state
// concurrently to other go routines mutating it

// some functions/fields/methods that might be helpful
// txo.PrsTXOLoc(...)
// bc.Lock()
// bc.Unlock()
func (bc *Blockchain) GetUTXOForAmt(amt uint32, pubKey string) ([]*UTXOInfo, uint32, bool) {
	bc.Lock()
	defer bc.Unlock()

	lastUTXO := bc.LastBlock.utxo

	// k, _ := json.Marshal(lastUTXO)
	// fmt.Println(string(k))
	// fmt.Printf("Amt Requested: %d\nFor Pubkey: %s\n", amt, pubKey)
	var availableUTXO uint32 = 0
	utxoForTransaction := []*UTXOInfo{}
	var change uint32 = 0
	if amt == 0 {
		return utxoForTransaction, 0, true
	}

	for key, output := range lastUTXO {
		// this is payable to the pubkey
		if output.LockingScript == pubKey {
			fmt.Printf("utxo found with amt %d\n", output.Amount)
			txHash, txIndex := txo.PrsTXOLoc(key)
			newInfo := &UTXOInfo{
				TxHsh:  txHash,
				OutIdx: txIndex,
				UTXO:   output,
				Amt:    output.Amount,
			}
			utxoForTransaction = append(utxoForTransaction, newInfo)
			availableUTXO += output.Amount
		}
		if availableUTXO >= amt {
			// we have found enough UTXO to fill the transaction
			change = availableUTXO - amt
			return utxoForTransaction, change, true
		}
	}
	fmt.Println("not enough for tx")
	return utxoForTransaction, 0, false
}

// GenesisBlock creates the genesis block from
// the configuration of the blockchain. Most
// of the values in the genesis block are
// hardcoded in.
// Inputs:
// conf *Config the configuration of the block
// chain
// Returns:
// *block.Block the genesis block
func GenesisBlock(conf *Config) *block.Block {
	txoo := []*proto.TransactionOutput{proto.NewTxOutpt(conf.InitSbsdy, conf.GenPK)}
	genTx := proto.NewTx(0, nil, txoo, 0)
	return block.Deserialize(&proto.Block{
		Header: &proto.BlockHeader{
			Version:          0,
			PrevBlockHash:    "",
			MerkleRoot:       "",
			Timestamp:        0,
			DifficultyTarget: "",
			Nonce:            0,
		},
		Transactions: []*proto.Transaction{genTx},
	})
}

// GetBalance gets the balance for a particular person
// on the network.
// Inputs:
// pk string the public key of the person whose balance
// is trying to be identified represented as a serialized
// hex string
// Returns:
// uint32 the balance that the person has
func (bc *Blockchain) GetBalance(pk string) uint32 {
	var bal uint32 = 0
	for _, v := range bc.LastBlock.utxo {
		if v.LockingScript == pk {
			bal += v.Amount
		}
	}
	return bal
}

func (bc *Blockchain) String() string {
	bc.Lock()
	defer bc.Unlock()
	b := bc.LastBlock
	slice := make([]string, 0)
	for ct := bc.LastBlock.depth + 1; ct > 0; ct-- {
		slice = append([]string{b.Block.NameTag()}, slice...)
		b = b.PrevNode
	}
	return fmt.Sprintf("[%v]", strings.Join(slice, ", "))
}
