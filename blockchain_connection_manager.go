package main

import (
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"google.golang.org/grpc"
)

type BlockchainConnection struct {
	clientConnection *grpc.ClientConn
	gateway          *client.Gateway
	network          *client.Network
}

func NewLayer1ConnectionManager() *BlockchainConnection {
	clientConnection := newGrpcConnection()

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

	network := gateway.GetNetwork("l1")

	return &BlockchainConnection{clientConnection: clientConnection, gateway: gateway, network: network}
}

func NewLayer2ConnectionManager() *BlockchainConnection {
	clientConnection := newGrpcConnection()

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

	network := gateway.GetNetwork("l2")

	return &BlockchainConnection{clientConnection: clientConnection, gateway: gateway, network: network}
}
