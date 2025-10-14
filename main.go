package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/network/api/http"
	"github.com/andantan/modular-blockchain/network/consensus/pbft"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/network/protocol/tcp"
	"github.com/andantan/modular-blockchain/network/provider/dns"
	"github.com/andantan/modular-blockchain/network/synchronizer"
	"github.com/andantan/modular-blockchain/server"
	"github.com/andantan/modular-blockchain/util"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func randomPortInRange(min, max uint16) (uint16, error) {
	if min >= max {
		return 0, fmt.Errorf("invalid range")
	}
	var b [2]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	r := binary.BigEndian.Uint16(b[:])
	port := min + (r % (max - min + 1))
	return port, nil
}

func main() {
	privKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		panic(err)
	}
	pubKey := privKey.PublicKey()
	address := pubKey.Address()

	blockChain := core.NewBlockchain()
	mpCap := 3000
	memPool := core.NewMemPool(mpCap)

	listenPort, err := randomPortInRange(49152, 65535)
	listenAddr := fmt.Sprintf("127.0.0.1:%d", listenPort)
	maxPeer := 8
	node := tcp.NewTcpNode(privKey, listenAddr, maxPeer)

	peerProviderAddr := "127.0.0.1:26550"
	peerProvider := dns.NewDnsPeerProvider(peerProviderAddr)

	apiListenAddr := fmt.Sprintf("127.0.0.1:%d", listenPort+1)
	apiServer := http.NewHttpApiServer(apiListenAddr, blockChain, memPool)

	storer := core.NewBlockStorage(fmt.Sprintf("%d_block", listenPort))
	blockTime := 30 * time.Second
	processor := core.NewBlockProcessor(blockChain)

	blockChain.SetBlockProcessor(processor)
	blockChain.SetStorer(storer)

	syncer := synchronizer.NewChainSynchronizer(address, listenAddr, blockChain)
	syncMsgCodec := synchronizer.NewDefaultSyncMessageCodec()
	gossipMsgCodec := message.NewDefaultGossipMessageCodec()

	no := server.NewNetworkOptions()
	no = no.WithNode(listenAddr, maxPeer, node).WithApiServer(apiListenAddr, apiServer).WithPeerProvider(peerProvider)
	no = no.WithGossipMessageCodec(gossipMsgCodec).WithSynchronizer(syncer).WithSyncMessageCodec(syncMsgCodec)

	bo := server.NewBlockchainOptions()
	bo = bo.WithStorer(storer).WithChain(blockChain, blockTime).WithVirtualMemoryPool(memPool, mpCap).WithProcessor(processor)

	consensusProposer := pbft.NewPbftProposer(privKey, blockChain, memPool)
	consensusLeaderSelector := pbft.NewPbftLeaderSelector()
	consensusEngineFactory := pbft.NewPbftConsensusEngineFactory()
	consensusMsgCodec := pbft.NewPbftConsensusMessageCodec()

	co := server.NewConsensusOptions()
	co = co.WithProposer(consensusProposer).WithLeaderSelector(consensusLeaderSelector)
	co = co.WithConsensusEngineFactory(consensusEngineFactory).WithConsensusMessageCodec(consensusMsgCodec)

	logger := util.LoggerWithPrefixes("Server")

	so := server.NewServerOptions(true)
	so = so.WithLogger(logger).WithPrivateKey(privKey)

	s := server.NewServer(so, no, bo, co)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go s.Start(wg)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	s.Stop()

	wg.Wait()
}
