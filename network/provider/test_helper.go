package provider

import (
	"fmt"
	"github.com/andantan/modular-blockchain/network/message"
	"github.com/andantan/modular-blockchain/util"
	"math/rand"
	"testing"
)

func GenerateRandomPeerInfo(t *testing.T) *message.PeerInfo {
	t.Helper()

	return &message.PeerInfo{
		Address:        util.RandomHash().String(),
		NetAddr:        fmt.Sprintf("127.0.0.1:%d", 10000+rand.Intn(50000)),
		Connections:    uint8(rand.Intn(8)),
		MaxConnections: 8,
		Height:         uint64(rand.Intn(1000)),
		IsValidator:    rand.Intn(2) == 0,
	}
}
