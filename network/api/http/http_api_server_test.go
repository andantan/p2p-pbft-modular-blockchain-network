package http

import (
	"bytes"
	"encoding/json"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHttpApiServer_HandleCreateTransaction(t *testing.T) {
	apiServer := NewHttpApiServer(":3000", nil, nil)
	apiServer.Start()

	t.Cleanup(func() {
		apiServer.Stop()
	})

	t.Run("it_should_process_a_valid_transaction", func(t *testing.T) {
		txReq := GenerateTestRandomHttpTransactionCreateRequest(t)
		origTx, err := txReq.ToTransaction()
		assert.NoError(t, err)
		origHash, err := origTx.Hash()
		assert.NoError(t, err)

		body, err := json.Marshal(txReq)
		assert.NoError(t, err)

		req := httptest.NewRequest("POST", "/create/transaction", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		apiServer.handleCreateTransaction(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		select {
		case receivedTx := <-apiServer.ConsumeTransaction():
			recvHash, err := receivedTx.Hash()
			assert.NoError(t, err)
			assert.True(t, origHash.Equal(recvHash))
		case <-time.After(50 * time.Millisecond):
			t.Fatal("API server did not send the transaction to the channel")
		}
	})

	// --- 3. 실패 시나리오: 잘못된 서명 ---
	t.Run("it_should_reject_a_transaction_with_invalid_signature", func(t *testing.T) {
		txReq := GenerateTestRandomHttpTransactionCreateRequest(t)
		sigTemper := &crypto.Signature{}
		txReq.Signature = sigTemper.String()

		body, _ := json.Marshal(txReq)
		req := httptest.NewRequest("POST", "/create/transaction", bytes.NewReader(body))
		rec := httptest.NewRecorder()

		apiServer.handleCreateTransaction(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
