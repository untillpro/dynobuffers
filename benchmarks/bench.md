# Optimization 2020-07-30

## Before

goos: windows
goarch: amd64
pkg: github.com/untillpro/dynobuffers/benchmarks
BenchmarkWriteDynoBuffersSimpleTyped-8                           1000000              1220 ns/op             456 B/op         14 allocs/op
BenchmarkWriteDynoBuffersSimpleTypedReadWriteString-8            1000000              1259 ns/op             456 B/op         14 allocs/op
BenchmarkWriteDynoBuffersSimpleUntyped-8                         1000000              1235 ns/op             464 B/op         16 allocs/op
BenchmarkWriteFlatBuffersSimple-8                                1937850               730 ns/op             256 B/op         10 allocs/op
BenchmarkWriteJsonSimple-8                                        307854              5767 ns/op            1352 B/op         34 allocs/op
BenchmarkWriteLinkedInAvroSimple-8                               1000000              1715 ns/op             776 B/op         13 allocs/op
BenchmarkWriteDynoBuffersArticleReadFewFieldsTyped-8             6125574               225 ns/op              80 B/op          1 allocs/op
BenchmarkWriteFlatBuffersArticleReadFewFields-8                   226525              7906 ns/op            3640 B/op         18 allocs/op
BenchmarkWriteJsonArticleReadFewFields-8                            7128            157240 ns/op           26396 B/op        856 allocs/op
BenchmarkWriteLinkedInAvroArticleReadFewFields-8                   38277             31411 ns/op           11929 B/op        176 allocs/op
BenchmarkWriteFlatBuffersArticleReadAllFields-8                   107749              9602 ns/op            3640 B/op         18 allocs/op
BenchmarkWriteJsonArticleReadAllFields-8                            7504            184782 ns/op           29467 B/op        856 allocs/op
BenchmarkWriteDynoBufferArticleReadAllFieldsUntyped-8              54084             20727 ns/op            7040 B/op        105 allocs/op
BenchmarkWriteDynoBufferArticleReadAllFieldsTyped-8                80538             16217 ns/op            6312 B/op         22 allocs/op
BenchmarkWriteSimple_Dyno_SameBuilder-8                          2049012               517 ns/op             144 B/op          4 allocs/op
BenchmarkWriteSimple_Dyno-8                                      1000000              1682 ns/op             520 B/op         16 allocs/op
BenchmarkWriteSimple_Avro-8                                      2481793               665 ns/op             344 B/op          3 allocs/op
PASS
ok      github.com/untillpro/dynobuffers/benchmarks     28.160s

## After


goos: windows
goarch: amd64
pkg: github.com/Yohanson555/dynobuffers/benchmarks
BenchmarkWriteDynoBuffersSimpleTyped-8                           3218420               346 ns/op               0 B/op          0 allocs/op
BenchmarkWriteDynoBuffersSimpleTypedReadWriteString-8            3363159               352 ns/op               0 B/op          0 allocs/op
BenchmarkWriteDynoBuffersSimpleUntyped-8                         3062869               401 ns/op               8 B/op          2 allocs/op
BenchmarkWriteFlatBuffersSimple-8                                1542766               747 ns/op             256 B/op         10 allocs/op
BenchmarkWriteJsonSimple-8                                        315969              4948 ns/op            1352 B/op         34 allocs/op
BenchmarkWriteLinkedInAvroSimple-8                                704853              1707 ns/op             776 B/op         13 allocs/op
BenchmarkWriteDynoBuffersArticleReadFewFieldsTyped-8             9026899               121 ns/op               0 B/op          0 allocs/op
BenchmarkWriteFlatBuffersArticleReadFewFields-8                   255538              5566 ns/op            3640 B/op         18 allocs/op
BenchmarkWriteJsonArticleReadFewFields-8                            7995            158022 ns/op           26398 B/op        856 allocs/op
BenchmarkWriteLinkedInAvroArticleReadFewFields-8                   45699             27025 ns/op           11930 B/op        176 allocs/op
BenchmarkWriteFlatBuffersArticleReadAllFields-8                   146401              8580 ns/op            3640 B/op         18 allocs/op
BenchmarkWriteJsonArticleReadAllFields-8                            7998            168216 ns/op           29463 B/op        856 allocs/op
BenchmarkWriteNestedSimple_Dyno-8                                 631964              1907 ns/op               0 B/op          0 allocs/op
BenchmarkWriteNestedSimple_Dyno_SameBuilder-8                     631953              1877 ns/op               0 B/op          0 allocs/op
BenchmarkWriteNestedSimple_ApplyMap_Test-8                       1352124               892 ns/op               0 B/op          0 allocs/op
BenchmarkWriteNested_ToBytes_Test-8                              1283980               933 ns/op               0 B/op          0 allocs/op
BenchmarkWriteNested_ToBytes_Parallel-8                          4366150               287 ns/op               0 B/op          0 allocs/op
BenchmarkWriteDynoBufferArticleReadAllFieldsUntyped-8             100485             12821 ns/op             728 B/op         83 allocs/op
BenchmarkWriteDynoBufferArticleReadAllFieldsTyped-8                87640             16907 ns/op            4988 B/op         23 allocs/op
BenchmarkWriteSimple_Dyno_SameBuilder-8                          4111824               248 ns/op               0 B/op          0 allocs/op
BenchmarkWriteSimple_Dyno-8                                      3845793               314 ns/op               0 B/op          0 allocs/op
BenchmarkWriteSimple_Avro-8                                      2122617               762 ns/op             344 B/op          3 allocs/op
PASS
ok      github.com/Yohanson555/dynobuffers/benchmarks   36.061s
