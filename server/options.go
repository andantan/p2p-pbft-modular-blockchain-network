package server

import (
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/api"
	"github.com/andantan/modular-blockchain/network/consensus"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/network/protocol"
	"github.com/andantan/modular-blockchain/network/provider"
	"github.com/andantan/modular-blockchain/network/synchronizer"
	"time"
)

type NetworkOptions struct {
	MaxPeers           int
	Node               protocol.Node
	Provider           provider.PeerProvider
	ApiServer          api.ApiServer
	GossipMessageCodec message.GossipMessageCodec
	Synchronizer       synchronizer.Synchronizer
	SyncMessageCodec   synchronizer.SyncMessageCodec
}

func NewNetworkOptions() *NetworkOptions {
	return &NetworkOptions{}
}

func (o *NetworkOptions) WithNode(maxPeers int, node protocol.Node) *NetworkOptions {
	o.MaxPeers = maxPeers
	o.Node = node
	return o
}

func (o *NetworkOptions) WithProvider(p provider.PeerProvider) *NetworkOptions {
	o.Provider = p
	return o
}

func (o *NetworkOptions) WithApiServer(s api.ApiServer) *NetworkOptions {
	o.ApiServer = s
	return o
}

func (o *NetworkOptions) WithGossipMessageCodec(c message.GossipMessageCodec) *NetworkOptions {
	o.GossipMessageCodec = c
	return o
}

func (o *NetworkOptions) WithSynchronizer(s synchronizer.Synchronizer) *NetworkOptions {
	o.Synchronizer = s
	return o
}

func (o *NetworkOptions) WithSyncMessageCodec(c synchronizer.SyncMessageCodec) *NetworkOptions {
	o.SyncMessageCodec = c
	return o
}

func (o *NetworkOptions) IsFulFilled() bool {
	return o.MaxPeers > 0 && o.Node != nil && o.ApiServer != nil && o.Provider == nil && o.GossipMessageCodec != nil && o.Synchronizer != nil && o.SyncMessageCodec != nil
}

type BlockchainOptions struct {
	Storer             core.Storer
	Chain              core.Chain
	BlockTime          time.Duration
	MemoryPool         core.VirtualMemoryPool
	MemoryPoolCapacity int
	Processor          core.Processor
}

func NewBlockchainOptions() *BlockchainOptions {
	return &BlockchainOptions{}
}

func (o *BlockchainOptions) WithStorer(s core.Storer) *BlockchainOptions {
	o.Storer = s
	return o
}

func (o *BlockchainOptions) WithChain(chain core.Chain, blockTime time.Duration) *BlockchainOptions {
	o.Chain = chain
	o.BlockTime = blockTime
	return o
}

func (o *BlockchainOptions) WithMemoryPool(memoryPool core.VirtualMemoryPool, capacity int) *BlockchainOptions {
	o.MemoryPool = memoryPool
	o.MemoryPoolCapacity = capacity
	return o
}

func (o *BlockchainOptions) WithProcessor(p core.Processor) *BlockchainOptions {
	o.Processor = p
	return o
}

func (o *BlockchainOptions) IsFulFilled() bool {
	return o.Storer != nil && o.Chain != nil && o.BlockTime > 0 && o.MemoryPool != nil && o.MemoryPoolCapacity != 0 && o.Processor != nil
}

type ConsensusOptions struct {
	Proposer              consensus.Proposer
	ConsensusEngines      map[uint64]consensus.ConsensusEngine
	ConsensusMessageCodec consensus.ConsensusMessageCodec
}

func NewConsensusOptions() *ConsensusOptions {
	return &ConsensusOptions{
		ConsensusEngines: make(map[uint64]consensus.ConsensusEngine),
	}
}

func (o *ConsensusOptions) WithProposer(p consensus.Proposer) *ConsensusOptions {
	o.Proposer = p
	return o
}

func (o *ConsensusOptions) WithConsensusMessageCodec(c consensus.ConsensusMessageCodec) *ConsensusOptions {
	o.ConsensusMessageCodec = c
	return o
}

func (o *ConsensusOptions) IsFulFilled() bool {
	return o.Proposer != nil && o.ConsensusEngines != nil && o.ConsensusMessageCodec != nil
}

type ServerOptions struct {
	PrivateKey    *crypto.PrivateKey
	ListenAddr    string
	ApiListenAddr string
	IsValidator   bool
}

func NewServerOptions(isValidator bool) *ServerOptions {
	return &ServerOptions{
		IsValidator: isValidator,
	}
}

func (o *ServerOptions) WithPrivateKey(k *crypto.PrivateKey) *ServerOptions {
	o.PrivateKey = k
	return o
}

func (o *ServerOptions) WithListenAddr(listenAddr string) *ServerOptions {
	o.ListenAddr = listenAddr
	return o
}

func (o *ServerOptions) WithApiListenAddr(apiListenAddr string) *ServerOptions {
	o.ApiListenAddr = apiListenAddr
	return o
}

func (o *ServerOptions) IsFulFilled() bool {
	return o.PrivateKey != nil && o.ListenAddr != "" && o.ApiListenAddr != ""
}
