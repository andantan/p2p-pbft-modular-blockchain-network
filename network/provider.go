package network

type PeerInfo struct {
	Address        string `json:"address"`
	NetAddr        string `json:"net_addr"`
	Connections    uint8  `json:"connections"`
	MaxConnections uint8  `json:"max_connections"`
	Height         uint64 `json:"height"`
	IsValidator    bool   `json:"is_validator"`
}

func NewPeerInfo(
	address string,
	netAddr string,
	connections uint8,
	maxConnections uint8,
	height uint64,
	isValidator bool,
) *PeerInfo {
	return &PeerInfo{
		Address:        address,
		NetAddr:        netAddr,
		Connections:    connections,
		MaxConnections: maxConnections,
		Height:         height,
		IsValidator:    isValidator,
	}
}

type PeerProvider interface {
	DiscoverPeers() ([]PeerInfo, error)
	GetValidators() ([]PeerInfo, error)
	Register(*PeerInfo) error
	Deregister(*PeerInfo) error
	Heartbeat(*PeerInfo) error
}
