package server

import (
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/api"
	"github.com/andantan/modular-blockchain/network/consensus"
	"github.com/andantan/modular-blockchain/network/protocol"
	"github.com/andantan/modular-blockchain/network/provider"
	"time"
)

type NetworkOptions struct {
	MaxPeers  int
	Node      protocol.Node
	Provider  provider.PeerProvider
	ApiServer api.ApiServer
}

func NewNetworkOptions() *NetworkOptions {
	return &NetworkOptions{}
}

func (o *NetworkOptions) WithNode(maxPeers int, node protocol.Node) *NetworkOptions {
	o.MaxPeers = maxPeers
	o.Node = node
	return o
}

func (o *NetworkOptions) WithProvider(provider provider.PeerProvider) *NetworkOptions {
	o.Provider = provider
	return o
}

func (o *NetworkOptions) WithApiServer(apiServer api.ApiServer) *NetworkOptions {
	o.ApiServer = apiServer
	return o
}

type BlockchainOptions struct {
	Storer             core.Storer
	Chain              core.Chain
	BlockTime          time.Duration
	MemoryPool         core.VirtualMemoryPool
	MemoryPoolCapacity int
	Processor          core.Processor
}

type ConsensusOptions struct {
	Proposer consensus.Proposer
	Engine   consensus.ConsensusEngine
}

type ServerOptions struct {
	PrivateKey    *crypto.PrivateKey
	ListenAddr    string
	ApiListenAddr string
}
