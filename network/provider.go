package network

import "github.com/andantan/modular-blockchain/network/message"

type PeerProvider interface {
	DiscoverPeers() ([]*message.PeerInfo, error)
	GetValidators() ([]*message.PeerInfo, error)
	Register(*message.PeerInfo) error
	Deregister(*message.PeerInfo) error
	Heartbeat(*message.PeerInfo) error
}
