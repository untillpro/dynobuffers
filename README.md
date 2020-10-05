[![codecov](https://codecov.io/gh/untillpro/dynobuffers/branch/master/graph/badge.svg)](https://codecov.io/gh/untillpro/dynobuffers)

# Dyno Buffers

- Codegen-less wrapper for [FlatBuffers](https://github.com/google/flatbuffers) with get\set by name feature
- Status: under development

# Features
- Uses FlatBuffer to read\write values from\to byte array converting to the required type described in the Scheme.
- No codegen, no compilers, no (de)serialization. Just fields description and get\set by name.
- In contrast to FlatBuffers tracks if the field was unset or initially not set
- Supported types
  - `int32, int64, float32, float64, bool, string, byte`
  - nested objects
  - arrays
- Scheme versioning
  - Any data written with Scheme of any version will be correctly read using Scheme of any other version
    - Written in old Scheme, read in New Scheme -> nil result on new field read, field considered as unset
    - Written in new Scheme, read in old Scheme -> no errors

# Limitations
- Only 2 cases of scheme modification are allowed: field rename and append fields to the end. This is necessary to have ability to read byte buffers in Scheme of any version
- Written in New -> read in Old -> write in Old -> New fields are lost (todo)

# Installation
`go get github.com/untillpro/dynobuffers`

# Usage
- Describe Scheme
  - By yaml. Field types:
    - `int` -> `int32`
    - `long` -> `int64`
    - `float` -> `float32`
    - `double` -> `float64`
    - `bool` -> `bool`
    - `string` -> `string`
    - `byte` -> `byte`
	```go
	var schemeStr = `
	name: string
	price: float
	quantity: int
	Id: long # first letter is capital -> field is mandatory
	`
	scheme, err := dynobuffers.YamlToScheme(schemeStr)
	if err != nil {
		panic(err)
	}
	```
  - Manually
	```go
	scheme := dynobuffers.NewScheme()
	scheme.AddField("name", dynobuffers.FieldTypeString, false)
	scheme.AddField("price", dynobuffers.FieldTypeFloat, false)
	scheme.AddField("quantity", dynobuffers.FieldTypeInt, false)
	scheme.AddField("id", dynobuffers.FieldTypeLong, true)
	```
- Create empty Dyno Buffer using Scheme
	```go
	b := dynobuffers.NewBuffer(scheme)
	```
	- panics if nil provided
- Set\modify fields according to the Scheme
	```go
	b.Set("price", float32(0.123))
	b.Set("name", nil) // unset field
	```
- To bytes array
	```go
	bytes := b.ToBytes()
	```
- To JSON key-value
    ```go
    jsonStr := b.ToJSON()
	```
	Note: arrays of byte are encoded to base64 strings
- To map key-value (JSON-compatible)
	```go
	jsonMap := b.ToJSONMap()
	```
- Read Buffer from bytes using Scheme
	```go
	b = dynobuffers.ReadBuffer(bytes, scheme)
	```
	- panics if nil Scheme provided
- Work with Buffer 
	```go
	value, ok := b.GetFloat32("price") // read typed. !ok -> field is unset or no such field in the scheme. Works faster and takes less memory allocations than Get()
	b.Get("price") // read untyped. nil -> field is unset or no such field in the scheme
	b.Set("price", nil) // set to nil means unset
	bytes = b.ToBytes()
	```
- Load data from JSON key-value and to bytes array
  	```go
	bytes, err := b.ApplyJSONAndToBytes([]byte(`{"name": "str", "price": 0.123, "fld": null}`))
	if err != nil {
		panic(err)
	}
	```
	- value is nil and field is mandatory -> error
	- value type and field type are incompatible (e.g. string provided for numeric field) -> error
    - value type and field type differs but value fits into field (e.g. float64(255) fits into float, double, int, long, byte; float64(256) does not fit into byte etc) -> ok
    - no such field in the scheme -> error
    - array element value is nil -> error (not supported)
- Check if a field exists in the scheme and is set to non-nil
  ```go
  b.HasValue("name")
  ```
- See [dynobuffers_test.go](dynobuffers_test.go) for usage examples

## Nested objects
- Declare scheme
  - by yaml
	```go
	var schemeStr = `
	name: string
	Nested: 
	  nes1: int
	  Nes2: int
	Id: long
	`
	schemeRoot := dynobuffers.YamlToScheme(schemeStr)
  - manually
	```go
	schemeNested := dynobuffers.NewScheme()
	schemeNested.AddField("nes1", dynobuffers.FieldTypeInt, false)
	schemeNested.AddField("nes2", dynobuffers.FieldTypeInt, true)
	schemeRoot := dynobuffers.NewScheme()
	schemeRoot.AddField("name", dynobuffers.FieldTypeString, false)
	schemeRoot.AddNested("nested", schemeNested, true)
	schemeRoot.AddField("id", dynobuffers.FieldTypeLong, true)
	```
- Create Buffer, fill and to bytes
	```go
	bRoot := dynobuffers.NewBuffer(schemeRoot)
	bNested := dynobuffers.NewBuffer(schemeNested)
	bNested.Set("nes1", 1)
	bNested.Set("nes2", 2)
	bRoot.Set("name", "str")
	bRoot.Set("nested", bNested)
	bytes, err := bRoot.ToBytes()
	if err != nil {
		panic(err)
	}
	```
- Read from bytes, modify and to bytes again
	```go
	bRoot = dynobuffers.ReadBuffer(bytes, schemeRoot)
	bRoot.Set("name", "str modified")
	bNested := bRoot.Get("nested").(*Buffer)
	bNested.Set("nes2", 3)
	bytes, err := bRoot.ToBytes()
	if err != nil {
		panic(err)
	}
	bRoot = dynobuffers.ReadBuffer(bytes, scheme)
	// note: bNested is obsolete here. Need to obtain it again from bRoot
	```
 - See [dynobuffers_test.go](dynobuffers_test.go) for usage examples

## Arrays
- Declare scheme
  - by yaml. Append `..` to field name to make it array
	```go
	var schemeStr = `
	name: string
	Nested..: 
	  nes1: int
	  Nes2: int
	Ids..: long
	`
	schemeRoot := dynobuffers.YamlToScheme(schemeStr)
  - manually
	```go
	schemeNested := dynobuffers.NewScheme()
	schemeNested.AddField("nes1", dynobuffers.FieldTypeInt, false)
	schemeNested.AddField("nes2", dynobuffers.FieldTypeInt, true)
	schemeRoot := dynobuffers.NewScheme()
	schemeRoot.AddField("name", dynobuffers.FieldTypeString, false)
	schemeRoot.AddNestedArray("nested", schemeNested, true)
	schemeRoot.AddArray("ids", dynobuffers.FieldTypeLong, true)
	```
- Create Buffer, fill and to bytes
	```go
	bRoot := dynobuffers.NewBuffer(schemeRoot)
	buffersNested := make([]*Buffer, 2)

	bNested := dynobuffers.NewBuffer(schemeNested)
	bNested.Set("nes1", 1)
	bNested.Set("nes2", 2)
	buffersNested = append(buffersNested, bNested)

	bNested = dynobuffers.NewBuffer(schemeNested)
	bNested.Set("nes1", 3)
	bNested.Set("nes2", 4)
	buffersNested = append(buffersNested, bNested)

	ids := []int64{5,6}
	bRoot.Set("name", "str")
	bRoot.Set("nested", buffersNested)
	bRoot.Set("ids", ids)
	bytes, err := bRoot.ToBytes()
	if err != nil {
		panic(err)
	}
	```
- Read array
  - By index
	```go
	bRoot = dynobuffers.ReadBuffer(bytes, schemeRoot)
	assert.Equal(t, int64(5), bRoot.GetByIndex("ids", 0))
	assert.Equal(t, int32(1), bRoot.GetByIndex("nested", 0).(*Buffer).Get("nes1"))
	```
  - Read filled array of non-objects
    ```go
	bRoot = dynobuffers.ReadBuffer(bytes, schemeRoot)
	arr := b.Get("ids").([]int64)
	```	
  - Read array of objects using iterator
    ```go
	bRoot = dynobuffers.ReadBuffer(bytes, schemeRoot)
	arr := bRoot.Get("nested").(*dynobuffers.ObjectArray)
	for arr.Next() {
		// arr.Buffer is switched on each arr.Next()
		assert.Equal(t, int32(1), arr.Buffer.Get("nes1"))
	}
	```
- Modify array and to bytes
	- Set
		```go
		bRoot = dynobuffers.ReadBuffer(bytes, schemeRoot)
		
		ids := []int64{5,6}
		bRoot.Set("ids", ids)

		arr := bRoot.Get("nested").(*dynobuffers.ObjectArray)
		arr.Next()
		arr.Buffer.Set("nes1", -1)
		bytes, err := bRoot.ToBytes()
		if err != nil {
			panic(err)
		}
		bRoot = dynobuffers.ReadBuffer(bytes, scheme)
		// note: elements of `buffers` array are obsolete here. Need to obtain the array again from bRoot
		```
	- Append
		```go
		bRoot = dynobuffers.ReadBuffer(bytes, schemeRoot)
		bRoot.Append("ids", []int32{7,8}) // if wasn't set then equivalent to Set()
		bytes, err := bRoot.ToBytes()
		if err != nil {
			panic(err)
		}
		```	
 - Nils as array elements are not supported
 - Byte arrays are decoded to JSON as base64 strings
 - Byte array value could be set from either byte array and base64-encoded string
 - See [dynobuffers_test.go](dynobuffers_test.go) for usage examples	

# Benchmarks

## Reading Many Fields

- cd benchmarks
- go test -bench=ReadAllArticleFields -benchmem

```
goos: windows
goarch: amd64
pkg: github.com/untillpro/dynobuffers/benchmarks
Benchmark_ReadAllArticleFields_Avro-8                      71046             23394 ns/op           11257 B/op        149 allocs/op
Benchmark_ReadAllArticleFields_Dyno_Untyped-8             195801              6437 ns/op             808 B/op         84 allocs/op
Benchmark_ReadAllArticleFields_Dyno_Typed-8               313376              3776 ns/op               0 B/op          0 allocs/op
Benchmark_ReadAllArticleFields_Flat-8                    1000000              1132 ns/op               0 B/op          0 allocs/op
Benchmark_ReadAllArticleFields_Json-8                      14431             87331 ns/op           14145 B/op        603 allocs/op
```

## Reading Few of Many Fields

- cd benchmarks
- go test -bench=ReadFewArticleFields -benchmem

```
goos: windows
goarch: amd64
pkg: github.com/untillpro/dynobuffers/benchmarks
Benchmark_ReadFewArticleFields_Avro-8              98437             19311 ns/op           11257 B/op        149 allocs/op
Benchmark_ReadFewArticleFields_Dyno_Typed-8     18520500                62.2 ns/op             0 B/op          0 allocs/op
Benchmark_ReadFewArticleFields_Flat-8           60032416                19.8 ns/op             0 B/op          0 allocs/op
Benchmark_ReadFewArticleFields_Json-8              15333             83824 ns/op           11073 B/op        603 allocs/op
```

## Reading Few Fields

- cd benchmarks
- go test -bench=ReadSimple -benchmem

```
goos: windows
goarch: amd64
pkg: github.com/untillpro/dynobuffers/benchmarks
Benchmark_ReadSimple_Avro-8              2038466              1193 ns/op             744 B/op         10 allocs/op
Benchmark_ReadSimple_Dyno_Typed-8       20017480                55.2 ns/op             0 B/op          0 allocs/op
Benchmark_ReadSimple_Flat-8             59981404                20.2 ns/op             0 B/op          0 allocs/op
Benchmark_ReadSimple_Flat_String-8      29275360                40.7 ns/op             0 B/op          0 allocs/op
Benchmark_ReadSimple_Json-8               545769              2423 ns/op             776 B/op         21 allocs/op
```

NOTE: DynoBuffers allocs caused by string types
