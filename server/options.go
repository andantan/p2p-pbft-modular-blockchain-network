package server

import (
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/api"
	"github.com/andantan/modular-blockchain/network/consensus"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/network/protocol"
	"github.com/andantan/modular-blockchain/network/provider"
	"github.com/andantan/modular-blockchain/network/synchronizer"
	"github.com/andantan/modular-blockchain/types"
	"time"
)

type NetworkOptions struct {
	MaxPeers           int
	Node               protocol.Node
	PeerProvider       provider.PeerProvider
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

func (o *NetworkOptions) WithPeerProvider(p provider.PeerProvider) *NetworkOptions {
	o.PeerProvider = p
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
	if o.MaxPeers <= 0 {
		return false
	}

	if o.Node == nil {
		return false
	}

	if o.ApiServer == nil {
		return false
	}

	if o.PeerProvider == nil {
		return false
	}

	if o.GossipMessageCodec == nil {
		return false
	}

	if o.Synchronizer == nil {
		return false
	}

	if o.SyncMessageCodec == nil {
		return false
	}

	return true
}

type BlockchainOptions struct {
	Storer            core.Storer
	Chain             core.Chain
	BlockTime         time.Duration
	VirtualMemoryPool core.VirtualMemoryPool
	Capacity          int
	Processor         core.Processor
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

func (o *BlockchainOptions) WithVirtualMemoryPool(mp core.VirtualMemoryPool, capacity int) *BlockchainOptions {
	o.VirtualMemoryPool = mp
	o.Capacity = capacity
	return o
}

func (o *BlockchainOptions) WithProcessor(p core.Processor) *BlockchainOptions {
	o.Processor = p
	return o
}

func (o *BlockchainOptions) IsFulFilled() bool {
	if o.Storer == nil {
		return false
	}

	if o.Chain == nil {
		return false
	}

	if o.BlockTime <= 0 {
		return false
	}

	if o.VirtualMemoryPool == nil {
		return false
	}

	if o.Capacity <= 0 {
		return false
	}

	if o.Processor == nil {
		return false
	}

	return true
}

type ConsensusOptions struct {
	Proposer                        consensus.Proposer
	ConsensusEngines                map[uint64]consensus.ConsensusEngine
	ConsensusEngineFactory          consensus.ConsensusEngineFactory
	ConsensusMessageCodec           consensus.ConsensusMessageCodec
	ForwardConsensusEngineMessageCh chan consensus.ConsensusMessage
	ForwardFinalizedBlockCh         chan *block.Block
}

func NewConsensusOptions() *ConsensusOptions {
	return &ConsensusOptions{
		ConsensusEngines:                make(map[uint64]consensus.ConsensusEngine),
		ForwardConsensusEngineMessageCh: make(chan consensus.ConsensusMessage, 200),
		ForwardFinalizedBlockCh:         make(chan *block.Block, 1),
	}
}

func (o *ConsensusOptions) WithProposer(p consensus.Proposer) *ConsensusOptions {
	o.Proposer = p
	return o
}

func (o *ConsensusOptions) WithConsensusEngineFactory(f consensus.ConsensusEngineFactory) *ConsensusOptions {
	o.ConsensusEngineFactory = f
	return o
}

func (o *ConsensusOptions) WithConsensusMessageCodec(c consensus.ConsensusMessageCodec) *ConsensusOptions {
	o.ConsensusMessageCodec = c
	return o
}

func (o *ConsensusOptions) IsFulFilled() bool {
	if o.Proposer == nil {
		return false
	}

	if o.ConsensusEngines == nil {
		return false
	}

	if o.ConsensusEngineFactory == nil {
		return false
	}

	if o.ConsensusMessageCodec == nil {
		return false
	}

	if o.ForwardConsensusEngineMessageCh == nil {
		return false
	}

	if o.ForwardFinalizedBlockCh == nil {
		return false
	}

	return true
}

type ServerOptions struct {
	PrivateKey    *crypto.PrivateKey
	PublicKey     *crypto.PublicKey
	Address       types.Address
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
	o.PublicKey = k.PublicKey()
	o.Address = k.PublicKey().Address()
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
	if o.PrivateKey == nil {
		return false
	}

	if o.PublicKey == nil {
		return false
	}

	if o.Address.Equal(types.Address{}) {
		return false
	}

	if o.ListenAddr == "" {
		return false
	}

	if o.ApiListenAddr == "" {
		return false
	}

	return true
}
