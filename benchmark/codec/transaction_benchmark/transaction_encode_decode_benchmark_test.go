package transaction_benchmark

import (
	"bytes"
	"encoding/gob"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pbBenchmark "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/benchmark/codec/transaction_benchmark"
	pbBlock "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"google.golang.org/protobuf/proto"
	"testing"
)

var (
	txx5000    = generateTransactions(5000)
	txx30000   = generateTransactions(30000)
	txx200000  = generateTransactions(200000)
	txx1000000 = generateTransactions(1000000)

	protoTxx5000    = transactionsToProto(txx5000)
	protoTxx30000   = transactionsToProto(txx30000)
	protoTxx200000  = transactionsToProto(txx200000)
	protoTxx1000000 = transactionsToProto(txx1000000)
)

func generateTransactions(n int) []*block.Transaction {
	txx := make([]*block.Transaction, n)
	for i := 0; i < n; i++ {
		tx := block.NewTransaction(util.RandomBytes(1<<10), uint64(i))
		privKey, _ := crypto.GeneratePrivateKey()
		_ = tx.Sign(privKey)
		txx[i] = tx
	}
	return txx
}

// ===== Gob benchmark =====

func benchmarkGobEncode(txx []*block.Transaction, b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(txx)
	}
}

func benchmarkGobDecode(txx []*block.Transaction, b *testing.B) {
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(txx)
	data := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decodedTxx []*block.Transaction
		_ = gob.NewDecoder(bytes.NewReader(data)).Decode(&decodedTxx)
	}
}

func BenchmarkGobEncode_5000(b *testing.B) {
	benchmarkGobEncode(txx5000, b)
}
func BenchmarkGobDecode_5000(b *testing.B) {
	benchmarkGobDecode(txx5000, b)
}
func BenchmarkGobEncode_30000(b *testing.B) {
	benchmarkGobEncode(txx30000, b)
}
func BenchmarkGobDecode_30000(b *testing.B) {
	benchmarkGobDecode(txx30000, b)
}
func BenchmarkGobEncode_200000(b *testing.B) {
	benchmarkGobEncode(txx200000, b)
}
func BenchmarkGobDecode_200000(b *testing.B) {
	benchmarkGobDecode(txx200000, b)
}
func BenchmarkGobEncode_1000000(b *testing.B) {
	benchmarkGobEncode(txx1000000, b)
}
func BenchmarkGobDecode_1000000(b *testing.B) {
	benchmarkGobDecode(txx1000000, b)
}

// ===== Protobuf benchmark =====

func transactionsToProto(txx []*block.Transaction) *pbBenchmark.Transactions {
	protoTxx := make([]*pbBlock.Transaction, len(txx))
	for i, tx := range txx {
		protoTx, _ := tx.ToProto()
		protoTxx[i] = protoTx.(*pbBlock.Transaction)
	}
	return &pbBenchmark.Transactions{Txx: protoTxx}
}

func benchmarkProtoEncode(protoTxx *pbBenchmark.Transactions, b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(protoTxx)
	}
}

func benchmarkProtoDecode(protoTxx *pbBenchmark.Transactions, b *testing.B) {
	data, _ := proto.Marshal(protoTxx)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decodedProtoTxx pbBenchmark.Transactions
		_ = proto.Unmarshal(data, &decodedProtoTxx)
	}
}

func BenchmarkProtoEncode_5000(b *testing.B) {
	benchmarkProtoEncode(protoTxx5000, b)
}
func BenchmarkProtoDecode_5000(b *testing.B) {
	benchmarkProtoDecode(protoTxx5000, b)
}
func BenchmarkProtoEncode_30000(b *testing.B) {
	benchmarkProtoEncode(protoTxx30000, b)
}
func BenchmarkProtoDecode_30000(b *testing.B) {
	benchmarkProtoDecode(protoTxx30000, b)
}
func BenchmarkProtoEncode_200000(b *testing.B) {
	benchmarkProtoEncode(protoTxx200000, b)
}
func BenchmarkProtoDecode_200000(b *testing.B) {
	benchmarkProtoDecode(protoTxx200000, b)
}
func BenchmarkProtoEncode_1000000(b *testing.B) {
	benchmarkProtoEncode(protoTxx1000000, b)
}
func BenchmarkProtoDecode_1000000(b *testing.B) {
	benchmarkProtoDecode(protoTxx1000000, b)
}
