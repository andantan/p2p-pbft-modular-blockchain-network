package provider

type PeerInfo struct {
	Address        string `json:"address"`
	NetAddr        string `json:"net_addr"`
	Connections    uint8  `json:"connections"`
	MaxConnections uint8  `json:"max_connections"`
	Height         uint64 `json:"height"`
	IsValidator    bool   `json:"is_validator"`
}

type PeerProvider interface {
	DiscoverPeers() ([]*PeerInfo, error)
	GetValidators() ([]*PeerInfo, error)
	Register(*PeerInfo) error
	Deregister(*PeerInfo) error
	Heartbeat(*PeerInfo) error
}
