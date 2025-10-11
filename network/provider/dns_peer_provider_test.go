package provider

import (
	"encoding/json"
	"github.com/andantan/modular-blockchain/network/message"
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

		var receivedInfo message.PeerInfo
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&receivedInfo))
		assert.Equal(t, i.Address, receivedInfo.Address)
		assert.Equal(t, i.NetAddr, receivedInfo.NetAddr)
		assert.Equal(t, i.Connections, receivedInfo.Connections)
		assert.Equal(t, i.MaxConnections, receivedInfo.MaxConnections)
		assert.Equal(t, i.Height, receivedInfo.Height)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := NewDnsPeerProvider(server.Listener.Addr().String())
	assert.NoError(t, provider.Register(i))
}

func TestDnsPeerProvider_Deregister(t *testing.T) {
	i := GenerateRandomPeerInfo(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/deregister", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var receivedInfo message.PeerInfo
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&receivedInfo))
		assert.Equal(t, i.Address, receivedInfo.Address)
		assert.Equal(t, i.NetAddr, receivedInfo.NetAddr)
		assert.Equal(t, i.Connections, receivedInfo.Connections)
		assert.Equal(t, i.MaxConnections, receivedInfo.MaxConnections)
		assert.Equal(t, i.Height, receivedInfo.Height)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := NewDnsPeerProvider(server.Listener.Addr().String())
	assert.NoError(t, provider.Deregister(i))
}

func TestDnsPeerProvider_Heartbeat(t *testing.T) {
	i := GenerateRandomPeerInfo(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/heartbeat", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var receivedInfo message.PeerInfo
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&receivedInfo))
		assert.Equal(t, i.Address, receivedInfo.Address)
		assert.Equal(t, i.NetAddr, receivedInfo.NetAddr)
		assert.Equal(t, i.Connections, receivedInfo.Connections)
		assert.Equal(t, i.MaxConnections, receivedInfo.MaxConnections)
		assert.Equal(t, i.Height, receivedInfo.Height)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := NewDnsPeerProvider(server.Listener.Addr().String())
	assert.NoError(t, provider.Heartbeat(i))
}

func TestDnsPeerProvider_DiscoverPeers(t *testing.T) {
	peerN := 50
	expectedPeers := make([]*message.PeerInfo, peerN)
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

	provider := NewDnsPeerProvider(server.Listener.Addr().String())
	peers, err := provider.DiscoverPeers()

	assert.NoError(t, err)
	assert.Equal(t, expectedPeers, peers)
}

func TestDnsPeerProvider_GetValidators(t *testing.T) {
	peerN := 50
	expectedPeers := make([]*message.PeerInfo, peerN)
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

	provider := NewDnsPeerProvider(server.Listener.Addr().String())
	validators, err := provider.GetValidators()

	assert.NoError(t, err)
	assert.Equal(t, expectedPeers, validators)
}
