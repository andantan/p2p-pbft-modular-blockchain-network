package dns

import (
	"encoding/json"
	"github.com/andantan/modular-blockchain/network/provider"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDnsPeerProvider_Register(t *testing.T) {
	i := GenerateRandomPeerInfo(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/register", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var receivedInfo provider.PeerInfo
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&receivedInfo))
		assert.Equal(t, i.Address, receivedInfo.Address)
		assert.Equal(t, i.NetAddr, receivedInfo.NetAddr)
		assert.Equal(t, i.Connections, receivedInfo.Connections)
		assert.Equal(t, i.MaxConnections, receivedInfo.MaxConnections)
		assert.Equal(t, i.Height, receivedInfo.Height)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := NewDnsPeerProvider(server.Listener.Addr().String())
	assert.NoError(t, p.Register(i))
}

func TestDnsPeerProvider_Deregister(t *testing.T) {
	i := GenerateRandomPeerInfo(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/deregister", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var receivedInfo provider.PeerInfo
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&receivedInfo))
		assert.Equal(t, i.Address, receivedInfo.Address)
		assert.Equal(t, i.NetAddr, receivedInfo.NetAddr)
		assert.Equal(t, i.Connections, receivedInfo.Connections)
		assert.Equal(t, i.MaxConnections, receivedInfo.MaxConnections)
		assert.Equal(t, i.Height, receivedInfo.Height)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := NewDnsPeerProvider(server.Listener.Addr().String())
	p.Deregister(i)
}

func TestDnsPeerProvider_Heartbeat(t *testing.T) {
	i := GenerateRandomPeerInfo(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/heartbeat", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var receivedInfo provider.PeerInfo
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&receivedInfo))
		assert.Equal(t, i.Address, receivedInfo.Address)
		assert.Equal(t, i.NetAddr, receivedInfo.NetAddr)
		assert.Equal(t, i.Connections, receivedInfo.Connections)
		assert.Equal(t, i.MaxConnections, receivedInfo.MaxConnections)
		assert.Equal(t, i.Height, receivedInfo.Height)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	p := NewDnsPeerProvider(server.Listener.Addr().String())
	assert.NoError(t, p.Heartbeat(i))
}

func TestDnsPeerProvider_DiscoverPeers(t *testing.T) {
	peerN := 50
	expectedPeers := make([]*provider.PeerInfo, peerN)
	for i := 0; i < peerN; i++ {
		peer := GenerateRandomPeerInfo(t)
		expectedPeers[i] = peer
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/peers", r.URL.Path)

		resp := PeerSliceResponse{Peers: expectedPeers}
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	p := NewDnsPeerProvider(server.Listener.Addr().String())
	peers, err := p.DiscoverPeers()

	assert.NoError(t, err)
	assert.Equal(t, expectedPeers, peers)
}

func TestDnsPeerProvider_GetValidators(t *testing.T) {
	peerN := 50
	expectedPeers := make([]*provider.PeerInfo, peerN)
	for i := 0; i < peerN; i++ {
		peer := GenerateRandomPeerInfo(t)
		expectedPeers[i] = peer
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/validators", r.URL.Path)

		resp := PeerSliceResponse{Peers: expectedPeers}
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(resp))
	}))
	defer server.Close()

	p := NewDnsPeerProvider(server.Listener.Addr().String())
	validators, err := p.GetValidators()

	assert.NoError(t, err)
	assert.Equal(t, expectedPeers, validators)
}
