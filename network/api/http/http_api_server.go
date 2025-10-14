package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/andantan/modular-blockchain/core"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	"github.com/andantan/modular-blockchain/util"
	"github.com/go-kit/log"
	"net/http"
	"sync"
	"time"
)

type HttpApiServer struct {
	logger     log.Logger
	listenAddr string
	server     *http.Server
	chain      core.Chain
	pool       core.VirtualMemoryPool

	txCh    chan *block.Transaction
	closeCh chan struct{}

	closeOnce sync.Once
}

func NewHttpApiServer(listenAddr string, chain core.Chain, pool core.VirtualMemoryPool) *HttpApiServer {
	mux := http.NewServeMux()

	s := &HttpApiServer{
		logger:     util.LoggerWithPrefixes("ApiServer"),
		listenAddr: listenAddr,
		server: &http.Server{
			Addr:    listenAddr,
			Handler: mux,
		},
		chain:   chain,
		pool:    pool,
		txCh:    make(chan *block.Transaction, 100),
		closeCh: make(chan struct{}),
	}

	mux.HandleFunc("/create/transaction", s.handleCreateTransaction)
	mux.HandleFunc("/random", s.handleMakeRandomTransaction)

	return s
}

func (s *HttpApiServer) Start() {
	_ = s.logger.Log("msg", "started api server", "addr", s.listenAddr)
	go func() {
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			panic("HttpApiServer: ListenAndServe: " + err.Error())
		}
	}()
}

func (s *HttpApiServer) ConsumeTransaction() <-chan *block.Transaction {
	return s.txCh
}

func (s *HttpApiServer) Stop() {
	s.closeOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		close(s.closeCh)
		close(s.txCh)

		_ = s.server.Shutdown(ctx)
		_ = s.logger.Log("msg", "shutdown http api server complete")
	})
}

func (s *HttpApiServer) handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	select {
	case <-s.closeCh:
		return
	default:
	}

	var txReq HttpTransactionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&txReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := txReq.ToTransaction()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = tx.Verify(); err != nil {
		_ = s.logger.Log("warn", "invalid transaction received", "err", err)
		http.Error(w, "invalid transaction signature", http.StatusBadRequest)
		return
	}

	txHash, err := tx.Hash()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	select {
	case <-s.closeCh:
		return
	case s.txCh <- tx:
	}

	resp := HttpTransactionCreateResponse{
		TransactionHash: txHash.String(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *HttpApiServer) handleMakeRandomTransaction(w http.ResponseWriter, r *http.Request) {
	privKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return
	}

	randomData := util.RandomBytes(512)
	tx := block.NewTransaction(randomData, util.RandomUint64())
	_ = tx.Sign(privKey)

	txHash, err := tx.Hash()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	select {
	case <-s.closeCh:
		return
	case s.txCh <- tx:
	}

	resp := HttpTransactionCreateResponse{
		TransactionHash: txHash.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
