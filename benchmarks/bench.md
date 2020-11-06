# Optimization 2020-07-30


## Before

```
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
```

## After

```
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
```

## After fix + parallel

```
goos: windows
goarch: amd64
pkg: github.com/untillpro/dynobuffers/benchmarks

Benchmark_RW_Pbill_Json-4                          	    4285	    259992 ns/op	  110828 B/op	    2258 allocs/op
Benchmark_RW_Simple_Dyno_Typed-4                   	12120525	       146 ns/op	       0 B/op	       0 allocs/op
Benchmark_RW_Simple_Dyno_Typed_String-4            	 9301798	       120 ns/op	       0 B/op	       0 allocs/op
Benchmark_RW_Simple_Dyno_Untyped-4                 	 9835501	       143 ns/op	       4 B/op	       1 allocs/op
Benchmark_RW_Simple_Flat-4                         	22640227	        46.7 ns/op	       0 B/op	       0 allocs/op
Benchmark_RW_Simple_Json-4                         	  599964	      2239 ns/op	    1360 B/op	      35 allocs/op
Benchmark_RW_Simple_Avro-4                         	 2708648	       474 ns/op	     416 B/op	       8 allocs/op
Benchmark_RW_Article_FewFields_Dyno_Typed-4        	  749956	      1610 ns/op	       0 B/op	       0 allocs/op
Benchmark_RW_Article_FewFields_Flat-4              	 1411683	       853 ns/op	       0 B/op	       0 allocs/op
Benchmark_RW_Article_FewFields_Json-4              	    9998	    100026 ns/op	   46663 B/op	     870 allocs/op
Benchmark_RW_Article_FewFields_Avro-4              	   80532	     13921 ns/op	   11154 B/op	      76 allocs/op
Benchmark_RW_Article_AllFields_Flat-4              	  999940	      1193 ns/op	       0 B/op	       0 allocs/op
Benchmark_RW_Article_AllFields_Json-4              	   12446	     96261 ns/op	   46597 B/op	     868 allocs/op
Benchmark_RW_Article_AllFields_Dyno_Untyped-4      	  352921	      3426 ns/op	     352 B/op	      25 allocs/op
Benchmark_RW_Article_AllFields_Dyno_Typed-4        	  461512	      3155 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_PbillItem_ByIndex-4                    	 2334496	       551 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_PBillItem_Iter-4                       	 3680769	       305 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_Simple_Dyno_Typed_String-4             	22472500	        58.7 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_Simple_Dyno_Untyped-4                  	20688477	        61.7 ns/op	       4 B/op	       1 allocs/op
Benchmark_R_Simple_Avro-4                          	 1951107	       703 ns/op	     728 B/op	       8 allocs/op
Benchmark_R_Simple_Dyno_Typed-4                    	41377027	        44.0 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_Simple_Flat-4                          	116488400	        11.0 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_Simple_Flat_String-4                   	89762998	        18.4 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_Simple_Json-4                          	 1077320	      1174 ns/op	     440 B/op	      19 allocs/op
Benchmark_R_Article_FewFields_Avro-4               	  101688	     10189 ns/op	   10650 B/op	      70 allocs/op
Benchmark_R_Article_FewFields_Dyno_Typed-4         	37497890	        39.0 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_Article_FewFields_Flat-4               	88333203	        12.9 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_Article_FewFields_Json-4               	   31166	     34142 ns/op	   11074 B/op	     603 allocs/op
Benchmark_R_Article_AllFields_Avro-4               	   99166	     11880 ns/op	   10650 B/op	      70 allocs/op
Benchmark_R_Article_AllFields_Dyno_Untyped-4       	  666625	      1956 ns/op	     352 B/op	      25 allocs/op
Benchmark_R_Article_AllFields_Dyno_Typed-4         	  923019	      1204 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_Article_AllFields_Flat-4               	 3508570	       337 ns/op	       0 B/op	       0 allocs/op
Benchmark_R_Article_AllFields_Json-4               	   32964	     38984 ns/op	   14146 B/op	     603 allocs/op
Benchmark_Fill_ToBytes_Simple_Dyno_SameBuilder-4   	10525714	       119 ns/op	       0 B/op	       0 allocs/op
Benchmark_Fill_ToBytes_Simple_Dyno-4               	 6976341	       147 ns/op	       0 B/op	       0 allocs/op
Benchmark_ApplyMap_Nested_Dyno-4                   	 2185666	       551 ns/op	       0 B/op	       0 allocs/op
Benchmark_MapToBytes_Nested_Dyno-4                 	  666628	      2175 ns/op	    1121 B/op	      10 allocs/op
Benchmark_MapToBytes_Nested_Dyno_SameBuilder-4     	  631545	      2193 ns/op	    1121 B/op	      10 allocs/op
Benchmark_MapToBytes_Pbill-4                       	   56178	     21754 ns/op	      49 B/op	       9 allocs/op
Benchmark_MapToBytes_PBill_Append-4                	    9229	    136967 ns/op	   70469 B/op	     451 allocs/op
Benchmark_MapToBytes_Simple_Avro-4                 	 4444189	       271 ns/op	     336 B/op	       2 allocs/op
Benchmark_MapToBytes_ArraysAppend_Dyno-4           	  799951	      1963 ns/op	     616 B/op	      30 allocs/op
Benchmark_MapToBytes_ArraysAppendNested_Dyno-4     	  517329	      3534 ns/op	    1014 B/op	      35 allocs/op
Benchmark_JSONUnmarshal_MapToBytes_Nested_Dyno-4   	  110084	     10329 ns/op	    6427 B/op	      84 allocs/op
Benchmark_JSONToBytes_Nested_Dyno-4                	  499970	      2622 ns/op	     416 B/op	      13 allocs/op
Benchmark_JSONToBytes_Simple_Dyno-4                	 4528042	       295 ns/op	      48 B/op	       3 allocs/op
Benchmark_JSONToBytes_Simple_Avro-4                	 1597777	       924 ns/op	     552 B/op	      19 allocs/op
Benchmark_Fill_ToBytes_Read_Simple_Dyno-4          	 4724139	       235 ns/op	       0 B/op	       0 allocs/op
Benchmark_ToJSON_Simple_Dyno-4                     	 5172006	       225 ns/op	     112 B/op	       2 allocs/op
```