package contract

import "github.com/andantan/modular-blockchain/core/block"

type Contract interface {
	Execute(transaction *block.Transaction) error
}
