package api

import "github.com/andantan/modular-blockchain/core/block"

type ApiServer interface {
	Start()
	ConsumeTransaction() <-chan *block.Transaction
	Stop()
}
