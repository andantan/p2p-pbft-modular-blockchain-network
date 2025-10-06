package block_benchmark

import (
	"bytes"
	"encoding/gob"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/crypto"
	pbBenchmark "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/benchmark/codec/block_benchmark"
	pbBlock "github.com/andantan/p2p-pbft-modular-blockchain-network/proto/core/block"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/types"
	"github.com/andantan/p2p-pbft-modular-blockchain-network/util"
	"google.golang.org/protobuf/proto"
	"testing"
)

var (
	bs1Txx5000    = generateBlocks(1, 5000)
	bs1Txx30000   = generateBlocks(1, 30000)
	bs1Txx200000  = generateBlocks(1, 200000)
	bs1Txx1000000 = generateBlocks(1, 1000000)

	bs10Txx5000   = generateBlocks(10, 5000)
	bs10Txx30000  = generateBlocks(10, 30000)
	bs10Txx200000 = generateBlocks(10, 200000)

	bs100Txx5000  = generateBlocks(100, 5000)
	bs100Txx30000 = generateBlocks(100, 30000)

	protoBs1Txx5000    = blocksToProto(bs1Txx5000)
	protoBs1Txx30000   = blocksToProto(bs1Txx30000)
	protoBs1Txx200000  = blocksToProto(bs1Txx200000)
	protoBs1Txx1000000 = blocksToProto(bs1Txx1000000)

	protoBs10Txx5000   = blocksToProto(bs10Txx5000)
	protoBs10Txx30000  = blocksToProto(bs10Txx30000)
	protoBs10Txx200000 = blocksToProto(bs10Txx200000)

	protoBs100Txx5000  = blocksToProto(bs100Txx5000)
	protoBs100Txx30000 = blocksToProto(bs100Txx30000)
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

func benchmarkGobEncodeBlocks(bs []*block.Block, b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := new(bytes.Buffer)
		_ = gob.NewEncoder(buf).Encode(bs)
	}
}

func benchmarkGobDecodeBlocks(bs []*block.Block, b *testing.B) {
	buf := new(bytes.Buffer)
	_ = gob.NewEncoder(buf).Encode(bs)
	data := buf.Bytes()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decodedBlocks []*block.Block
		_ = gob.NewDecoder(bytes.NewReader(data)).Decode(&decodedBlocks)
	}
}

func blocksToProto(blocks []*block.Block) *pbBenchmark.Blocks {
	protoBlocks := make([]*pbBlock.Block, len(blocks))

	for i, b := range blocks {
		protoBlock, _ := b.ToProto()
		protoBlocks[i] = protoBlock.(*pbBlock.Block)
	}

	return &pbBenchmark.Blocks{
		Blocks: protoBlocks,
	}
}

func benchmarkProtoEncodeBlocks(protoBlocks *pbBenchmark.Blocks, b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(protoBlocks)
	}
}

func benchmarkProtoDecodeBlocks(protoBlocks *pbBenchmark.Blocks, b *testing.B) {
	data, _ := proto.Marshal(protoBlocks)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var decodedProtoBlocks pbBenchmark.Blocks
		_ = proto.Unmarshal(data, &decodedProtoBlocks)
	}
}

// ===== Gob benchmark =====

func BenchmarkGobEncode_Blocks1_Txs5000(b *testing.B) {
	benchmarkGobEncodeBlocks(bs1Txx5000, b)
}
func BenchmarkGobDecode_Blocks1_Txs5000(b *testing.B) {
	benchmarkGobDecodeBlocks(bs1Txx5000, b)
}
func BenchmarkGobEncode_Blocks1_Txs30000(b *testing.B) {
	benchmarkGobEncodeBlocks(bs1Txx30000, b)
}
func BenchmarkGobDecode_Blocks1_Txs30000(b *testing.B) {
	benchmarkGobDecodeBlocks(bs1Txx30000, b)
}
func BenchmarkGobEncode_Blocks1_Txs200000(b *testing.B) {
	benchmarkGobEncodeBlocks(bs1Txx200000, b)
}
func BenchmarkGobDecode_Blocks1_Txs200000(b *testing.B) {
	benchmarkGobDecodeBlocks(bs1Txx200000, b)
}
func BenchmarkGobEncode_Blocks1_Txs1000000(b *testing.B) {
	benchmarkGobEncodeBlocks(bs1Txx1000000, b)
}
func BenchmarkGobDecode_Blocks1_Txs1000000(b *testing.B) {
	benchmarkGobDecodeBlocks(bs1Txx1000000, b)
}

func BenchmarkGobEncode_Blocks10_Txs5000(b *testing.B) {
	benchmarkGobEncodeBlocks(bs10Txx5000, b)
}
func BenchmarkGobDecode_Blocks10_Txs5000(b *testing.B) {
	benchmarkGobDecodeBlocks(bs10Txx5000, b)
}
func BenchmarkGobEncode_Blocks10_Txs30000(b *testing.B) {
	benchmarkGobEncodeBlocks(bs10Txx30000, b)
}
func BenchmarkGobDecode_Blocks10_Txs30000(b *testing.B) {
	benchmarkGobDecodeBlocks(bs10Txx30000, b)
}
func BenchmarkGobEncode_Blocks10_Txs200000(b *testing.B) {
	benchmarkGobEncodeBlocks(bs10Txx200000, b)
}
func BenchmarkGobDecode_Blocks10_Txs200000(b *testing.B) {
	benchmarkGobDecodeBlocks(bs10Txx200000, b)
}

func BenchmarkGobEncode_Blocks100_Txs5000(b *testing.B) {
	benchmarkGobEncodeBlocks(bs100Txx5000, b)
}
func BenchmarkGobDecode_Blocks100_Txs5000(b *testing.B) {
	benchmarkGobDecodeBlocks(bs100Txx5000, b)
}
func BenchmarkGobEncode_Blocks100_Txs30000(b *testing.B) {
	benchmarkGobEncodeBlocks(bs100Txx30000, b)
}
func BenchmarkGobDecode_Blocks100_Txs30000(b *testing.B) {
	benchmarkGobDecodeBlocks(bs100Txx30000, b)
}

// ===== Protobuf benchmark =====

func BenchmarkProtoEncode_Blocks1_Txs5000(b *testing.B) {
	benchmarkProtoEncodeBlocks(protoBs1Txx5000, b)
}
func BenchmarkProtoDecode_Blocks1_Txs5000(b *testing.B) {
	benchmarkProtoDecodeBlocks(protoBs1Txx5000, b)
}
func BenchmarkProtoEncode_Blocks1_Txs30000(b *testing.B) {
	benchmarkProtoEncodeBlocks(protoBs1Txx30000, b)
}
func BenchmarkProtoDecode_Blocks1_Txs30000(b *testing.B) {
	benchmarkProtoDecodeBlocks(protoBs1Txx30000, b)
}
func BenchmarkProtoEncode_Blocks1_Txs200000(b *testing.B) {
	benchmarkProtoEncodeBlocks(protoBs1Txx200000, b)
}
func BenchmarkProtoDecode_Blocks1_Txs200000(b *testing.B) {
	benchmarkProtoDecodeBlocks(protoBs1Txx200000, b)
}
func BenchmarkProtoEncode_Blocks1_Txs1000000(b *testing.B) {
	benchmarkProtoEncodeBlocks(protoBs1Txx1000000, b)
}
func BenchmarkProtoDecode_Blocks1_Txs1000000(b *testing.B) {
	benchmarkProtoDecodeBlocks(protoBs1Txx1000000, b)
}

func BenchmarkProtoEncode_Blocks10_Txs5000(b *testing.B) {
	benchmarkProtoEncodeBlocks(protoBs10Txx5000, b)
}
func BenchmarkProtoDecode_Blocks10_Txs5000(b *testing.B) {
	benchmarkProtoDecodeBlocks(protoBs10Txx5000, b)
}
func BenchmarkProtoEncode_Blocks10_Txs30000(b *testing.B) {
	benchmarkProtoEncodeBlocks(protoBs10Txx30000, b)
}
func BenchmarkProtoDecode_Blocks10_Txs30000(b *testing.B) {
	benchmarkProtoDecodeBlocks(protoBs10Txx30000, b)
}
func BenchmarkProtoEncode_Blocks10_Txs200000(b *testing.B) {
	benchmarkProtoEncodeBlocks(protoBs10Txx200000, b)
}
func BenchmarkProtoDecode_Blocks10_Txs200000(b *testing.B) {
	benchmarkProtoDecodeBlocks(protoBs10Txx200000, b)
}

func BenchmarkProtoEncode_Blocks100_Txs5000(b *testing.B) {
	benchmarkProtoEncodeBlocks(protoBs100Txx5000, b)
}
func BenchmarkProtoDecode_Blocks100_Txs5000(b *testing.B) {
	benchmarkProtoDecodeBlocks(protoBs100Txx5000, b)
}
func BenchmarkProtoEncode_Blocks100_Txs30000(b *testing.B) {
	benchmarkProtoEncodeBlocks(protoBs100Txx30000, b)
}
func BenchmarkProtoDecode_Blocks100_Txs30000(b *testing.B) {
	benchmarkProtoDecodeBlocks(protoBs100Txx30000, b)
}
