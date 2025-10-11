package http

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
)

type HttpTransactionCreateRequest struct {
	Data      []byte `json:"data"`
	From      string `json:"from"`
	Signature string `json:"signature"`
	Nonce     uint64 `json:"nonce"`
}

func (r *HttpTransactionCreateRequest) ToTransaction() (*block.Transaction, error) {
	pubKey, err := crypto.PublicKeyFromHexString(r.From)
	if err != nil {
		return nil, err
	}

	sig, err := crypto.SignatureFromHexString(r.Signature)
	if err != nil {
		return nil, err
	}

	return &block.Transaction{
		Data:      r.Data,
		From:      pubKey,
		Signature: sig,
		Nonce:     r.Nonce,
	}, nil
}
