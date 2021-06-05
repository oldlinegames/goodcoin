package test

import (
	"BrunoCoin/pkg"
	"BrunoCoin/pkg/block"
	"BrunoCoin/pkg/block/tx"
	"BrunoCoin/pkg/blockchain"
	"BrunoCoin/pkg/proto"
	"BrunoCoin/pkg/utils"
	"testing"
	"time"
)

func TestGetUTXOForAmt(t *testing.T) {
	utils.SetDebug(true)

	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Start()
	node2.Start()

	genNd.ConnectToPeer(node2.Addr)

	time.Sleep(1 * time.Second)

	genNd.SendTx(50, 100, node2.Id.GetPublicKeyBytes())

	time.Sleep(time.Second * 3)

	//not enough UTXO
	utxoForTransaction, change, wasEnough := genNd.Chain.GetUTXOForAmt(1000, node2.Id.GetPublicKeyBytes())

	if utxoForTransaction != nil {
		t.Fail()
	}
	if change != 0 {
		t.Fail()
	}
	if wasEnough {
		t.Fail()
	}

	utxoForTransaction, change, wasEnough := genNd.Chain.GetUTXOForAmt(10, node2.Id.GetPublicKeyBytes())

	if utxoForTransaction == nil {
		t.Fail()
	}
	if change != 40 {
		t.Fail()
	}
	if !wasEnough {
		t.Fail()
	}

}

func TestAdd(t *testing.T) {
	utils.SetDebug(true)

	genNd := NewGenNd()
	node2 := pkg.New(pkg.DefaultConfig(GetFreePort()))
	genNd.Start()
	node2.Start()

	genNd.ConnectToPeer(node2.Addr)

	time.Sleep(1 * time.Second)

	inputAmounts := 100
	input := &proto.TransactionInput{
		OutputIndex: 0,
		Amount:      100,
	}

	output := &proto.TransactionOutput{
		Amount:        100,
		LockingScript: 0,
	}
	transaction := &proto.Transaction{
		Inputs:   input,
		Outputs:  output,
	}

	transactions := tx.Deserialize(transaction)
	newblock := block.New(genNd.Chain.GetLastBlock().Hash(), transactions)
	//test for invalid block
	result := genNd.Chain.Add(newblock)

	if result != block {
		t.Fail()
	}
}