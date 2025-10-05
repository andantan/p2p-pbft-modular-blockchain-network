Benchmarks: Gob vs. Protocol Buffers
This benchmark compares the performance of Go's native gob encoding with Protocol Buffers (protobuf) for serializing and deserializing large slices of Transaction objects.

Results
Tests were run on an Intel(R) Core(TM) i5-1035G7 CPU @ 1.20GHz.

Benchmark (Transactions)	Library	Speed (ns/op)	Memory Usage (B/op)	Allocations (allocs/op)
Encode 5,000	Gob	9,821,608	-	-
Encode 5,000	Protobuf	2,588,184	-	-
Decode 5,000	Gob	7,674,553	13,320,337	65,222
Decode 5,000	Protobuf	3,666,268	6,474,809	20,017
Encode 30,000	Gob	88,521,328	-	-
Encode 30,000	Protobuf	14,699,608	-	-
Decode 30,000	Gob	90,782,558	164,235,492	390,226
Decode 30,000	Protobuf	24,472,020	39,221,692	120,024
Encode 200,000	Gob	549,716,350	-	-
Encode 200,000	Protobuf	94,580,846	-	-
Decode 200,000	Gob	551,518,800	1,375,447,032	2,600,237
Decode 200,000	Protobuf	133,067,488	261,900,508	800,033
Lower values are better.

## Analysis 🧐
The results clearly demonstrate that Protocol Buffers significantly outperforms Gob in all metrics, with the performance gap widening as the data size increases.

Speed (ns/op) 🚀

Encoding: Protobuf is ~4-6 times faster than Gob.

Decoding: Protobuf is ~2-4 times faster than Gob.

Reason: Protobuf uses highly optimized, pre-generated code for serialization, whereas Gob relies on Go's runtime reflection, which introduces significant overhead.

Memory Usage & Allocations (B/op, allocs/op) 📉

Protobuf uses approximately 5 times less memory and makes 3 times fewer memory allocations during decoding.

Reason: The Protobuf binary format is more compact. It uses numeric field tags instead of full field names, resulting in a smaller data footprint. Fewer allocations reduce pressure on the garbage collector (GC), leading to better overall application performance.

## Conclusion ✅
For a high-performance, distributed system like a blockchain, Protocol Buffers is the superior choice. While gob offers simplicity for Go-only projects, Protobuf provides critical advantages in speed, data size, and memory efficiency, making it the standard for performance-sensitive network applications.