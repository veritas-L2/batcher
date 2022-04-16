package main

import "fmt"

const BATCH_SIZE = 10

func getStateHash() {

}

func Batcher(transactionInfoBuffer chan TransactionInfo) {
	var currentBatch []TransactionInfo

	for {
		transactionInfo := <-transactionInfoBuffer
		currentBatch = append(currentBatch, transactionInfo)

		if len(currentBatch) == BATCH_SIZE {
			//publish batch to L1
			fmt.Printf(currentBatch[9].TransactionName)
			currentBatch = nil
		}
	}
}
