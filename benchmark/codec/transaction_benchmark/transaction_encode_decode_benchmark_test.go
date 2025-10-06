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

func transactionsToProto(txx []*block.Transaction) *pbBenchmark.Transactions {
	protoTxx := make([]*pbBlock.Transaction, len(txx))
	for i, tx := range txx {
		protoTx, _ := tx.ToProto()
		protoTxx[i] = protoTx.(*pbBlock.Transaction)
	}
	return &pbBenchmark.Transactions{Txx: protoTxx}
}

// ===== Gob benchmark =====

func BenchmarkGobEncode_5000(b *testing.B) {
	b.StopTimer()
	data := generateTransactions(5000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}

func BenchmarkGobDecode_5000(b *testing.B) {
	b.StopTimer()
	data := generateTransactions(5000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decodedTxx []*block.Transaction
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decodedTxx)
	}
}

func BenchmarkGobEncode_30000(b *testing.B) {
	b.StopTimer()
	data := generateTransactions(30000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}

func BenchmarkGobDecode_30000(b *testing.B) {
	b.StopTimer()
	data := generateTransactions(30000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decodedTxx []*block.Transaction
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decodedTxx)
	}
}

func BenchmarkGobEncode_200000(b *testing.B) {
	b.StopTimer()
	data := generateTransactions(200000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}

func BenchmarkGobDecode_200000(b *testing.B) {
	b.StopTimer()
	data := generateTransactions(200000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decodedTxx []*block.Transaction
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decodedTxx)
	}
}

func BenchmarkGobEncode_1000000(b *testing.B) {
	b.StopTimer()
	data := generateTransactions(1000000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}

func BenchmarkGobDecode_1000000(b *testing.B) {
	b.StopTimer()
	data := generateTransactions(1000000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decodedTxx []*block.Transaction
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decodedTxx)
	}
}

// ===== Protobuf benchmark =====

func BenchmarkProtoEncode_5000(b *testing.B) {
	b.StopTimer()
	data := transactionsToProto(generateTransactions(5000))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}

func BenchmarkProtoDecode_5000(b *testing.B) {
	b.StopTimer()
	data := transactionsToProto(generateTransactions(5000))
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Transactions
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}

func BenchmarkProtoEncode_30000(b *testing.B) {
	b.StopTimer()
	data := transactionsToProto(generateTransactions(30000))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}

func BenchmarkProtoDecode_30000(b *testing.B) {
	b.StopTimer()
	data := transactionsToProto(generateTransactions(30000))
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Transactions
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}

func BenchmarkProtoEncode_200000(b *testing.B) {
	b.StopTimer()
	data := transactionsToProto(generateTransactions(200000))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}

func BenchmarkProtoDecode_200000(b *testing.B) {
	b.StopTimer()
	data := transactionsToProto(generateTransactions(200000))
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Transactions
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}

func BenchmarkProtoEncode_1000000(b *testing.B) {
	b.StopTimer()
	data := transactionsToProto(generateTransactions(1000000))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}

func BenchmarkProtoDecode_1000000(b *testing.B) {
	b.StopTimer()
	data := transactionsToProto(generateTransactions(1000000))
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Transactions
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}
