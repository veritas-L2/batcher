package main

import (
	"encoding/json"
	"fmt"
	"time"
)

const BATCH_SIZE = 10

type Batch struct {
	Timestamp     int64             `json:"timestamp"`
	Transactions  []TransactionInfo `json:"transactions"`
	PrevStateHash []byte            `json:"prevStateHash"`
	NewStateHash  []byte            `json:"newStateHash"`
}

func newBatch() *Batch {
	return &Batch{Timestamp: 0, Transactions: []TransactionInfo{}, PrevStateHash: nil, NewStateHash: nil}
}

//TODO: handle errors
func (batch *Batch) CommitToLayer1(layer1Connection BlockchainConnection) {
	contract := layer1Connection.network.GetContract("veritas")
	json, _ := json.Marshal(batch)
	contract.SubmitTransaction("CommitBatch", string(json))
}

type Batcher struct {
	layer1Connection *BlockchainConnection
	layer2Connection *BlockchainConnection
	currentBatch     *Batch
	txnLock          bool
}

func NewBatcher(layer1Connection *BlockchainConnection, layer2Connection *BlockchainConnection) *Batcher {
	batch := newBatch()
	return &Batcher{layer1Connection: layer1Connection, layer2Connection: layer2Connection, currentBatch: batch}
}

func (batcher *Batcher) Run(transactionInfoBuffer chan TransactionInfo) {
	batcher.txnLock = true
	batcher.currentBatch.PrevStateHash = batcher.getStateHash()
	batcher.txnLock = false

	for {
		var transactionInfo TransactionInfo
		if batcher.txnLock == false {
			transactionInfo = <-transactionInfoBuffer
		} else {
			continue
		}
		contract := batcher.layer2Connection.network.GetContract(transactionInfo.ChaincodeName)
		result, err := contract.SubmitTransaction(transactionInfo.TransactionName, transactionInfo.Args...)
		if err != nil {
			fmt.Printf("txn failed to execute: %s\n", err.Error())
			continue
		} else {
			fmt.Printf("txn executed successfully. Result: %s\n", result)
		}

		batcher.currentBatch.Transactions = append(batcher.currentBatch.Transactions, transactionInfo)

		if len(batcher.currentBatch.Transactions) == BATCH_SIZE {
			//publish batch to L1
			batcher.currentBatch.Timestamp = time.Now().Unix()
			batcher.txnLock = true
			batcher.currentBatch.NewStateHash = batcher.getStateHash()
			batcher.txnLock = false

			fmt.Printf("%+v\n", batcher.currentBatch)

			newPrevStateHash := batcher.currentBatch.NewStateHash
			batcher.currentBatch = newBatch()
			batcher.currentBatch.PrevStateHash = newPrevStateHash
		}
	}
}

//TODO: handle error
func (batcher *Batcher) getStateHash() []byte {
	contract := batcher.layer2Connection.network.GetContract("state-contract")

	batcher.txnLock = true
	contract.SubmitTransaction("InitStateContract")
	result, _ := contract.SubmitTransaction("GetRootHash")
	contract.SubmitTransaction("ReleaseStateContract")
	batcher.txnLock = false
	return result
}
