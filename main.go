package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	org3MspID        = "Org3MSP"
	org3CryptoPath   = "/Users/ahsansyed/Desktop/workspace/veritas/fabric-veritas-samples/test-network/organizations/peerOrganizations/org3.example.com"
	org3CertPath     = org3CryptoPath + "/users/User1@org3.example.com/msp/signcerts/User1@org3.example.com-cert.pem"
	org3KeyPath      = org3CryptoPath + "/users/User1@org3.example.com/msp/keystore"
	org3TlsCertPath  = org3CryptoPath + "/peers/peer0.org3.example.com/tls/ca.crt"
	org3PeerEndpoint = "localhost:11051"
	org3GatewayPeer  = "peer0.org3.example.com"
	org3ChannelName  = "l2"
)

const (
	org1MspID        = "Org1MSP"
	org1CryptoPath   = "/Users/ahsansyed/Desktop/workspace/veritas/fabric-veritas-samples/test-network/organizations/peerOrganizations/org1.example.com"
	org1CertPath     = org1CryptoPath + "/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem"
	org1KeyPath      = org1CryptoPath + "/users/User1@org1.example.com/msp/keystore"
	org1TlsCertPath  = org1CryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	org1PeerEndpoint = "localhost:7051"
	org1GatewayPeer  = "peer0.org1.example.com"
	org1ChannelName  = "l1"
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
