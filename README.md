# Dyno Buffers

Codegen-less wrapper for [FlatBuffers](https://github.com/google/flatbuffers) with get\set by name feature

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
- Read Buffer from bytes using Scheme
	```go
	b = dynobuffers.ReadBuffer(bytes, scheme)
	```
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
    - no such field in the scheme -> ok (need to scheme versioning)
    - array element value is nil -> error (not supported)
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
  - Using `dynobuffers.Array` struct
    ```go
	bRoot = dynobuffers.ReadBuffer(bytes, schemeRoot)
	arr := b.Get("ids").(*dynobuffers.Array)
	```	
    - iterate
		```go
		for arr.Next() {
			value := arr.Value()
			fmt.Println(value)
		}
		```
	- get filled array as `interface{}` containing typed array
		```go
		filledArr := ar.GetAll()
		for i, value := range filledArr {
			fmt.Println(fmt.Sprintf("%d: %v", i, value)
		}
		```
	- get filled typed array 
		```go
		filledArr := ar.GetInts()
		for i, value := range filledArr {
			fmt.Println(fmt.Sprintf("%d: %d", i, value)
		}
		```
  - Fast read array of objects using `dynobuffers.ObjectArray` struct
    ```go
	bRoot = dynobuffers.ReadBuffer(bytes, schemeRoot)
	arr := dynobuffers.NewObjectArray()
	bRoot.GetArray("nested", arr)
	for arr.Next() {
		assert.Equal(t, int32(1), arr.Buffer.Get("nes1"))
	}
	```


- Modify array and to bytes
	- Set
		```go
		bRoot = dynobuffers.ReadBuffer(bytes, schemeRoot)
		
		ids := []int64{5,6}
		bRoot.Set("ids", ids)

		buffers := bRoot.Get("nested").(*dynobuffers.Array).GetObjects()
		buffers[1].Set("nes1", -1)
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
 - Byte arrays are decoded\encoded from\to JSON as base64 strings
 - See [dynobuffers_test.go](dynobuffers_test.go) for usage examples	

# Benchmarks
## Description
- [benchmarks\benchRead_test.go](benchmarks\benchRead_test.go) read benchmarks comparing to Avro, FlatBuffers, JSON
- [benchmarks\benchReadWrite_test.go](benchmarks\benchReadWrite_test.go) read and write benchmarks comparing to Avro, FlatBuffers, JSON
- Benchmarks naming
  - BenchmarkWrite... - benchmark read, change 1 field and write
  - BenchmarkRead... - benchmark read only
  - ...Simple... - scheme with 3 fields, read and multiply 2 fields
  - ...Article... - scheme with 123 fields
  - ...ReadFewFields... - read and multiply 2 fields
  - ...ReadAllFields... - read all fields
## Results
- Scheme of 123 fields, read and multiply 2 fields, change 1 field and write
```
DynoBuffers    	  	    200000	      8763 ns/op	    5240 B/op	      41 allocs/op
FlatBuffers         	    300000	      5702 ns/op	    3672 B/op	      19 allocs/op
Json                	     10000	    201575 ns/op	   29096 B/op	     857 allocs/op
LinkedInAvro                100000	     25173 ns/op	   11929 B/op	     176 allocs/op
```
- Scheme of 3 fields, read and multiply 2 fields
```
DynoBuffers               20000000	      65.0 ns/op	       0 B/op	       0 allocs/op
FlatBuffers               50000000	      33.2 ns/op	       0 B/op	       0 allocs/op
Json                        500000	      2660 ns/op	     760 B/op	      21 allocs/op
LinkedInAvro               2000000	       859 ns/op	     744 B/op	      10 allocs/op
```
- Scheme of 123 fields, read and multiple 2 fields
```
DynoBuffers     	  20000000	      83.2 ns/op	       0 B/op	       0 allocs/op
FlatBuffers          	 100000000	      26.5 ns/op	       0 B/op	       0 allocs/op
Json                 	     10000	    108663 ns/op	   13730 B/op	     604 allocs/op
LinkedInAvro         	    100000	     14936 ns/op	   11257 B/op	     149 allocs/op
```
- Scheme of 123 fields, simply read all fields
```
DynoBuffers      	    300000	      4935 ns/op	      35 B/op	      11 allocs/op
FlatBuffers          	   1000000	      1421 ns/op	       0 B/op	       0 allocs/op
Json                 	     10000	    116814 ns/op	   14130 B/op	     603 allocs/op
LinkedInAvro         	    100000	     19187 ns/op	   11257 B/op	     149 allocs/op
```

# To do
- Lists?
- Written in New -> read in Old -> write in Old -> New fields are kept.
