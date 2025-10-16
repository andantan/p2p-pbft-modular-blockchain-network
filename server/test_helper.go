package server

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/network/provider"
	"github.com/andantan/modular-blockchain/network/synchronizer"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
)

type MockPeerProvider struct {
	Logger log.Logger
	Peers  map[types.Address]*provider.PeerInfo
}

func GenerateMockPeerProvider() *MockPeerProvider {
	return &MockPeerProvider{
		Logger: util.LoggerWithPrefixes("MockPeerProvider"),
		Peers:  make(map[types.Address]*provider.PeerInfo),
	}
}

func (m *MockPeerProvider) Register(our *provider.PeerInfo) error {
	addr, _ := types.AddressFromHexString(our.Address)
	_ = m.Logger.Log("msg", "Register mock peer", "address", our.Address)
	m.Peers[addr] = our
	return nil
}

func (m *MockPeerProvider) DiscoverPeers() ([]*provider.PeerInfo, error) {
	_ = m.Logger.Log("msg", "Discover mock peer")
	ps := make([]*provider.PeerInfo, 0)
	for _, i := range m.Peers {
		ps = append(ps, i)
	}
	return ps, nil
}

func (m *MockPeerProvider) GetValidators() ([]*provider.PeerInfo, error) {
	_ = m.Logger.Log("msg", "GetValidators mock peer")
	vs := make([]*provider.PeerInfo, 0)
	for _, i := range m.Peers {
		if i.IsValidator {
			vs = append(vs, i)
		}
	}
	return vs, nil
}

func (m *MockPeerProvider) Heartbeat(our *provider.PeerInfo) error {
	_ = m.Logger.Log("msg", "Heartbeat mock peer", "address", our.Address)
	return nil
}

func (m *MockPeerProvider) Deregister(our *provider.PeerInfo) {
	_ = m.Logger.Log("msg", "Deregister mock peer", "address", our.Address)
	addr, _ := types.AddressFromHexString(our.Address)
	delete(m.Peers, addr)
	return
}

type MockApiServer struct {
	MockTxCh chan *block.Transaction
}

func GenerateMockApiServer() *MockApiServer {
	return &MockApiServer{}
}

func (m *MockApiServer) Start() {
	m.MockTxCh = make(chan *block.Transaction, 10)
}

func (m *MockApiServer) ConsumeTransaction() <-chan *block.Transaction {
	return m.MockTxCh
}

func (m *MockApiServer) Stop() {
	close(m.MockTxCh)
}

type MockSynchronizer struct{}

func GenerateMockSynchronizer() *MockSynchronizer {
	return &MockSynchronizer{}
}

func (m MockSynchronizer) Start() {
	panic("implement me")
}

func (m MockSynchronizer) HandleMessage(address types.Address, message synchronizer.SyncMessage) {
	panic("implement me")
}

func (m MockSynchronizer) OutgoingMessage() <-chan synchronizer.SynchronizerMessage {
	panic("implement me")
}

func (m MockSynchronizer) IsSynchronized() bool {
	return true
}

func (m MockSynchronizer) NotifyForkDetected(address types.Address, b *block.Block) {
	panic("implement me")
}

func (m MockSynchronizer) NotifyChainLagging() {
	panic("implement me")
}

func (m MockSynchronizer) Stop() {
	panic("implement me")
}
