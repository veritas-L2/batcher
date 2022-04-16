package main

// HELLA COPY PASTA COS WHY NOT

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

/*
step 1: receive txn
step 2: forward txn to HLF L2 channel
step 3: receive response from L2, forward response to client, add txn to batch
step 4: repeat until batch is complete or timeout
step 5: publish batch to L1 channel,
*/

const (
	mspID        = "Org3MSP"
	cryptoPath   = "/Users/ahsansyed/Desktop/workspace/veritas/fabric-veritas-samples/test-network/organizations/peerOrganizations/org3.example.com"
	certPath     = cryptoPath + "/users/User1@org3.example.com/msp/signcerts/User1@org3.example.com-cert.pem"
	keyPath      = cryptoPath + "/users/User1@org3.example.com/msp/keystore"
	tlsCertPath  = cryptoPath + "/peers/peer0.org3.example.com/tls/ca.crt"
	peerEndpoint = "localhost:11051"
	gatewayPeer  = "peer0.org3.example.com"
	channelName  = "l2"
)

type TransactionInfo struct {
	ChaincodeName   string   `form:"chaincodeName" json:"chaincodeName" xml:"chaincodeName"  binding:"required"`
	TransactionName string   `form:"transactionName" json:"transactionName" xml:"transactionName"  binding:"required"`
	Args            []string `form:"args" json:"args" xml:"args"  binding:"required"`
}

func main() {
	fmt.Println("hello world")

	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	network := gateway.GetNetwork(channelName)

	transactionBuffer := make(chan TransactionInfo)

	router := gin.Default()

	router.POST("/executeTransaction", func(c *gin.Context) {
		var transactionInfo TransactionInfo
		if c.BindJSON(&transactionInfo) == nil {
			contract := network.GetContract(transactionInfo.ChaincodeName)
			result, err := contract.SubmitTransaction(transactionInfo.TransactionName, transactionInfo.Args...)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": true, "response": err.Error()})
			} else {
				transactionBuffer <- transactionInfo
				c.JSON(http.StatusOK, gin.H{"error": false, "response": string(result)})
			}
		}
	})

	go Batcher(transactionBuffer)

	router.Run()
}
