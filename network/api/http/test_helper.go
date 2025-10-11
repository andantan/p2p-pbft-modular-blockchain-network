package http

import (
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func GenerateTestRandomHttpTransactionCreateRequest(t *testing.T) *HttpTransactionCreateRequest {
	t.Helper()

	privKey, _ := crypto.GeneratePrivateKey()
	tx := block.NewTransaction(util.RandomBytes(1<<10), util.RandomUint64())
	assert.NoError(t, tx.Sign(privKey))

	assert.NotNil(t, tx.From)
	assert.NotNil(t, tx.Signature)

	return &HttpTransactionCreateRequest{
		Data:      tx.Data,
		From:      tx.From.String(),
		Signature: tx.Signature.String(),
		Nonce:     tx.Nonce,
	}
}
