package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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

/*
step 1: receive txn
step 2: forward txn to HLF L2 channel
step 3: receive response from L2, forward response to client, add txn to batch
step 4: repeat until batch is complete or timeout
step 5: publish batch to L1 channel,
*/

type TransactionInfo struct {
	ChaincodeName   string   `form:"chaincodeName" json:"chaincodeName" xml:"chaincodeName"  binding:"required"`
	TransactionName string   `form:"transactionName" json:"transactionName" xml:"transactionName"  binding:"required"`
	Args            []string `form:"args" json:"args" xml:"args"  binding:"required"`
}

func main() {

	layer2Connection := NewLayer2ConnectionManager()
	layer1Connection := NewLayer1ConnectionManager()

	defer layer2Connection.clientConnection.Close()
	defer layer1Connection.clientConnection.Close()
	defer layer2Connection.gateway.Close()
	defer layer1Connection.gateway.Close()

	transactionBuffer := make(chan TransactionInfo)

	router := gin.Default()

	router.POST("/executeTransaction", func(c *gin.Context) {
		var transactionInfo TransactionInfo
		if c.BindJSON(&transactionInfo) == nil {
			transactionBuffer <- transactionInfo
			c.JSON(http.StatusOK, gin.H{"response": "transaction submitted"})
		}
	})

	batcher := NewBatcher(layer1Connection, layer2Connection)

	go batcher.Run(transactionBuffer)

	router.Run()
}
