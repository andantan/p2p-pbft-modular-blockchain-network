package protocol

import "github.com/andantan/modular-blockchain/types"

type Raw interface {
	Protocol() string
	From() types.Address
	Payload() []byte
}
