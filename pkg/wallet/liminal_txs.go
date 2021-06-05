package wallet

import (
	"BrunoCoin/pkg/block/tx"
	"sync"
)

/*
 *  Brown University, CS1951L, Summer 2021
 *  Designed by: Colby Anderson, Kotone Ninagawa
 */

// LiminalTxs (LiminalTransactions) are
// transactions that have been made by the
// wallet but are not considered valid yet.
// This is because the transactions either
// have not been mined to a block or the
// block doesn't have enough POW on top of it.
// TxQ is a maximum priority queue that stores
// transactions with priority being equivalent
// to how many blocks have been seen since the
// transaction was initially made.
// TxRplyThresh is the threshold priority for
// having blocks removed from the TxQ, sent out
// again, and added back with a priority of 0.
type LiminalTxs struct {
	TxQ          	*tx.Heap
	TxRplyThresh 	uint32
	mutex			sync.Mutex
}

// NewLmnlTxs (NewLiminalTransactions) returns
// a new Liminal Transactions object.
// Inputs:
// c *Config the configuration for the wallet
func NewLmnlTxs(c *Config) *LiminalTxs {
	return &LiminalTxs{
		TxQ:          tx.NewTxHeap(),
		TxRplyThresh: c.TxRplyThresh,
	}
}

// ChkTxs (CheckTransactions) checks that the inputted
// transactions from the new block aren't the same. This
// new block is assumed to have enough POW of work on top
// of it. If any transactions are the same, they are removed.
// Otherwise, the priorities are incremented, and any
// transaction with too large of a priority needs to be
// returned so it can be sent out again.
// Inputs:
// txs []*tx.Transaction a list of transactions that were in
// a valid block
// Returns:
//[]*tx.Transaction transactions with priorities above
// l.TxRplyThresh that are removed
//[]*tx.Transaction transactions from the new block that are already
//in LiminalTxs, so removed from LiminalTxs bc duplicates
// TODO
// 1. Remove duplicates
// 2. Increment all priorites
// 3. Remove transactions above a certain priority threshold
// Tip 1: Remember that this method will be
// mutating the LiminalTxs struct along with
// other go routines.

// some helpful functions/methods/fields:
// l.mutex.Lock()
// l.mutex.Unlock()
// l.TxQ.Rmv(...)
// l.TxQ.IncAll()
// l.TxQ.RemAbv(...)
func (l *LiminalTxs) ChkTxs(txs []*tx.Transaction) ([]*tx.Transaction, []*tx.Transaction) {
	// 1. Remove duplicates
	l.mutex.Lock()
	removedTransactions := l.TxQ.Rmv(txs)
	// Increment all priorties
	l.TxQ.IncAll()
	// 3. Remove transactions above a certain priority threshold
	aboveThres := l.TxQ.RemAbv(l.TxRplyThresh)
	l.mutex.Unlock()
	return aboveThres, removedTransactions
}


// Add adds a transaction to the liminal transactions.
// It is basically a wrapper around the heap add. The
// priority is 0, since the transaction was just made
// and no blocks have been retrieved since.
// Inputs:
// t *tx.Transaction the transaction to be added
// TODO
// Tip 1: Remember that this method will be
// mutating the LiminalTxs struct along with
// other go routines.

// some helpful functions/methods/fields:
// l.mutex.Lock()
// l.mutex.Unlock()
// l.TxQ.Add(...)
func (l *LiminalTxs) Add(t *tx.Transaction) {
	l.mutex.Lock()
	l.TxQ.Add(0, t)
	l.mutex.Unlock()
	return
}
