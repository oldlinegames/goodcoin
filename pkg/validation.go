package pkg

import (
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/block/tx/txi"
)

/*
 *  Brown University, CS1951L, Summer 2021
 *  Designed by: Colby Anderson, John Roy,
 *	Parker Ljung
 *
 */

// ChkBlk (CheckBlock) validates a block based on multiple
// conditions.
// To be valid:
// The block must be syntactically (ChkBlkSyn), semantically
// (ChkBlkSem), and configurally (ChkBlkConf) valid.
// Each transaction on the block must be syntactically (ChkTxSyn),
// semantically (ChkTxSem), and configurally (ChkTxConf) valid.
// Each transaction on the block must reference UTXO on the same
// chain (main or forked chain) and not be a double spend on that
// chain.
// Inputs:
// b *block.Block the block to be checked for validity
// Returns:
// bool True if the block is valid. false
// otherwise
// TODO:
// to be valid

// Each transaction on the block must reference UTXO on the same
// chain (main or forked chain) and not be a double spend on that
// chain.
// The block's size must be less than or equal to the largest
// allowed block size.
// The block hash must be less than the difficulty target.
// The block's first transaction must be of type Coinbase.

// Some helpful functions/methods/fields:
// note: let t be a transaction object
// note: let b be a block object
// t.IsCoinbase()
// b.SatisfiesPOW(...)
// n.Conf.MxBlkSz
// b.Sz()
// n.Chain.ChkChainsUTXO(...)
func (n *Node) ChkBlk(b *block.Block) bool {
	//Check node and block inputs
	if b == nil || n == nil {
		return false
	}
	// Check transactions
	if b.Transactions == nil || len(b.Transactions) == 0 {
		return false
	}
	// Verify that first tx is coinbase
	if !b.Transactions[0].IsCoinbase() {
		return false
	}
	// Verify hash is < difftarg
	if !b.SatisfiesPOW(b.Hdr.DiffTarg) {
		return false
	}
	// Check block size
	if b.Sz() > n.Conf.MxBlkSz {
		return false
	}
	// Check that all txs are referencing correct chain
	if !n.Chain.ChkChainsUTXO(b.Transactions, b.Hdr.PrvBlkHsh) {
		return false
	}
	// verify each transaction
	for _, newTx := range b.Transactions {
		if !n.ChkTx(newTx) {
			return false
		}
	}
	return true
}

// ChkTx (CheckTransaction) validates a transaction.
// Inputs:
// t *tx.Transaction the transaction to be checked for validity
// Returns:
// bool True if the transaction is syntactically valid. false
// otherwise
// TODO:
// to be valid:

// The transaction's inputs and outputs must not be empty.
// The transaction's output amounts must be larger than 0.
// The sum of the transaction's inputs must be larger
// than the sum of the transaction's outputs.
// The transaction must not double spend any UTXO.
// The unlocking script on each of the transaction's
// inputs must successfully unlock each of the corresponding
// UTXO.
// The transaction must not be larger than the
// maximum allowed block size.

// Some helpful functions/methods/fields:
// note: let t be a transaction object
// note: let b be a block object
// note: let u be a transaction output object
// n.Conf.MxBlkSz
// t.Sz()
// u.IsUnlckd(...)
// n.Chain.GetUTXO(...)
// n.Chain.IsInvalidInput(...)
// t.SumInputs()
// t.SumOutputs()
func (n *Node) ChkTx(t *tx.Transaction) bool {
	// Check that tx and its outputs are not nil or empty
	if t == nil || t.Inputs == nil || t.Outputs == nil {
		return false
	}
	if len(t.Inputs) == 0 || len(t.Outputs) == 0 {
		return false
	}
	// Check that tx inputs > outputs
	if !(t.SumInputs() > t.SumOutputs()) {
		return false
	}
	// Check size
	if t.Sz() > n.Conf.MxBlkSz {
		return false
	}
	// Check for double spending and verify locking scripts
	doubleSpendingCheckMap := make(map[*txi.TransactionInput]bool)
	for _, txInput := range t.Inputs {
		// check if its even valid
		if n.Chain.IsInvalidInput(txInput) {
			return false
		}
		// check for double spending
		if _, ok := doubleSpendingCheckMap[txInput]; ok {
			return false
		}
		// Check if unlocking script is good
		txOutput := n.Chain.GetUTXO(txInput)
		if !txOutput.IsUnlckd(txInput.UnlockingScript) {
			return false
		}
		doubleSpendingCheckMap[txInput] = true
	}
	return true
}
