package block

import (
	"bytes"
	"encoding/gob"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pb "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/block/transaction"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"google.golang.org/protobuf/proto"
	"testing"
)

var (
	txx5000   = generateTransactions(5000)
	txx30000  = generateTransactions(30000)
	txx200000 = generateTransactions(200000)

	protoTxx5000   = transactionsToProto(txx5000)
	protoTxx30000  = transactionsToProto(txx30000)
	protoTxx200000 = transactionsToProto(txx200000)
)

func generateTransactions(n int) []*Transaction {
	txx := make([]*Transaction, n)
	for i := 0; i < n; i++ {
		tx := NewTransaction(util.RandomBytes(1<<10), uint64(i))
		privKey, _ := crypto.GeneratePrivateKey()
		_ = tx.Sign(privKey)
		txx[i] = tx
	}
	return txx
}

// ===== Gob benchmark =====

func benchmarkGobEncode(txx []*Transaction, b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(txx)
	}
}

func benchmarkGobDecode(txx []*Transaction, b *testing.B) {
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(txx)
	data := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decodedTxx []*Transaction
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

// ===== Protobuf benchmark =====

func transactionsToProto(txx []*Transaction) *pb.Transactions {
	protoTxx := make([]*pb.Transaction, len(txx))
	for i, tx := range txx {
		protoTx, _ := tx.ToProto()
		protoTxx[i] = protoTx.(*pb.Transaction)
	}
	return &pb.Transactions{Txx: protoTxx}
}

func benchmarkProtoEncode(protoTxx *pb.Transactions, b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(protoTxx)
	}
}

func benchmarkProtoDecode(protoTxx *pb.Transactions, b *testing.B) {
	data, _ := proto.Marshal(protoTxx)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decodedProtoTxx pb.Transactions
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
