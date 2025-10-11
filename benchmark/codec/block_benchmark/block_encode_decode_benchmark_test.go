package block_benchmark

import (
	"bytes"
	"encoding/gob"
	"github.com/andantan/modular-blockchain/core/block"
	"github.com/andantan/modular-blockchain/crypto"
	pbBenchmark "github.com/andantan/modular-blockchain/proto/benchmark/codec/block_benchmark"
	pbBlock "github.com/andantan/modular-blockchain/proto/core/block"
	"github.com/andantan/modular-blockchain/types"
	"github.com/andantan/modular-blockchain/util"
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

func generateBlocks(bn int, txn int) []*block.Block {
	bs := make([]*block.Block, bn)

	prevHeader := &block.Header{
		Height:        0,
		PrevBlockHash: types.FFHash,
	}

	for i := 0; i < bn; i++ {
		k, _ := crypto.GeneratePrivateKey()
		txx := generateTransactions(txn)
		bd := block.NewBody(txx)
		b, _ := block.NewBlockFromPrevHeader(prevHeader, bd)
		_ = b.Sign(k)
		bs[i] = b
		prevHeader = b.Header
	}

	return bs
}

func generateProtoBlocks(bn int, txn int) *pbBenchmark.Blocks {
	blocks := generateBlocks(bn, txn)
	protoBlocks := make([]*pbBlock.Block, len(blocks))

	for i, b := range blocks {
		protoBlock, _ := b.ToProto()
		protoBlocks[i] = protoBlock.(*pbBlock.Block)
	}

	return &pbBenchmark.Blocks{
		Blocks: protoBlocks,
	}
}

// ===== Gob benchmark =====

func BenchmarkGobEncode_Blocks1_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(1, 5000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}
func BenchmarkGobDecode_Blocks1_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(1, 5000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decoded)
	}
}
func BenchmarkGobEncode_Blocks1_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(1, 30000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}
func BenchmarkGobDecode_Blocks1_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(1, 30000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decoded)
	}
}
func BenchmarkGobEncode_Blocks1_Txs200000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(1, 200000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}
func BenchmarkGobDecode_Blocks1_Txs200000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(1, 200000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decoded)
	}
}
func BenchmarkGobEncode_Blocks1_Txs1000000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(1, 1000000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}
func BenchmarkGobDecode_Blocks1_Txs1000000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(1, 1000000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decoded)
	}
}

func BenchmarkGobEncode_Blocks10_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(10, 5000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}
func BenchmarkGobDecode_Blocks10_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(10, 5000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decoded)
	}
}
func BenchmarkGobEncode_Blocks10_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(10, 30000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}
func BenchmarkGobDecode_Blocks10_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(10, 30000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decoded)
	}
}
func BenchmarkGobEncode_Blocks10_Txs200000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(10, 200000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}
func BenchmarkGobDecode_Blocks10_Txs200000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(10, 200000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decoded)
	}
}

func BenchmarkGobEncode_Blocks100_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(100, 5000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}
func BenchmarkGobDecode_Blocks100_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(100, 5000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decoded)
	}
}
func BenchmarkGobEncode_Blocks100_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(100, 30000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(data)
	}
}
func BenchmarkGobDecode_Blocks100_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateBlocks(100, 30000)
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(data)
	encodedData := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(encodedData)).Decode(&decoded)
	}
}

// ===== Protobuf benchmark =====

func BenchmarkProtoEncode_Blocks1_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(1, 5000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}
func BenchmarkProtoDecode_Blocks1_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(1, 5000)
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Blocks
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}
func BenchmarkProtoEncode_Blocks1_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(1, 30000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}
func BenchmarkProtoDecode_Blocks1_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(1, 30000)
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Blocks
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}
func BenchmarkProtoEncode_Blocks1_Txs200000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(1, 200000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}
func BenchmarkProtoDecode_Blocks1_Txs200000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(1, 200000)
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Blocks
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}
func BenchmarkProtoEncode_Blocks1_Txs1000000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(1, 1000000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}
func BenchmarkProtoDecode_Blocks1_Txs1000000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(1, 1000000)
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Blocks
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}

func BenchmarkProtoEncode_Blocks10_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(10, 5000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}
func BenchmarkProtoDecode_Blocks10_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(10, 5000)
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Blocks
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}
func BenchmarkProtoEncode_Blocks10_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(10, 30000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}
func BenchmarkProtoDecode_Blocks10_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(10, 30000)
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Blocks
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}
func BenchmarkProtoEncode_Blocks10_Txs200000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(10, 200000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}
func BenchmarkProtoDecode_Blocks10_Txs200000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(10, 200000)
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Blocks
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}

func BenchmarkProtoEncode_Blocks100_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(100, 5000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}
func BenchmarkProtoDecode_Blocks100_Txs5000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(100, 5000)
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Blocks
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}
func BenchmarkProtoEncode_Blocks100_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(100, 30000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(data)
	}
}
func BenchmarkProtoDecode_Blocks100_Txs30000(b *testing.B) {
	b.StopTimer()
	data := generateProtoBlocks(100, 30000)
	encodedData, _ := proto.Marshal(data)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded pbBenchmark.Blocks
		_ = proto.Unmarshal(encodedData, &decoded)
	}
}
