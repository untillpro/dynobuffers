# Dyno Buffers

Codegen-less wrapper for [FlatBuffers](https://github.com/google/flatbuffers) with get\set by name feature

# Features
- Uses FlatBuffer to read\write values from\to byte array converting to the required type described in the Schema.
- No codegen, no compilers, no (de)serialization. Just fields description and get\set by name.
- In contrast to FlatBuffers tracks if the field was unset or initially not set
- Supported types
  - `int32, int64, float32, float64, bool, string, byte`
- Schema versioning
  - Any data written with Schema of any version will be correctly read using Schema of any other version
    - Written in old Schema, read in New Schema -> nil result on new field read, field considered as unset
    - Written in new Schema, read in old Schema -> no errors

# Limitations
- Only 2 cases of schema modification are allowed: field rename and append fields to the end. This is necessary to have ability to read byte buffers in Schema of any version
- Written in New -> read in Old -> write in Old -> New fields are lost (todo)
- No arrays (todo)
- No nested objects (todo)
- No blobs (use strings?)?

# Installation
`go get github.com/untillpro/dynobuffers`

# Usage
- Describe Schema
  - By yaml. Field types:
    - `int` -> `int32`
    - `long` -> `int64`
    - `float` -> `float32`
    - `double` -> `float64`
    - `bool` -> `bool`
    - `string` -> `string`
    - `byte` -> `byte`
	```go
	var schemaStr = `
	name: string
	price: float
	quantity: int
	`
	schema, err := dynobuffers.YamlToSchema(schemaStr)
	if err != nil {
		panic(err)
	}
	```
  - Manually
	```go
	schema := dynobuffers.NewSchema()
	schema.AddField("name", dynobuffers.FieldTypeString)
	schema.AddField("price", dynobuffers.FieldTypeFloat)
	schema.AddField("quantity", dynobuffers.FieldTypeInt)
	```
- Create Dyno Buffer using Schema
	```go
	b := dynobuffers.NewBuffer(schema)
	```
- Set\modify fields according to the Schema
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
- Read Buffer from bytes using Schema
	```go
	b = dynobuffers.ReadBuffer(bytes, schema)
	```
- Work with Buffer 
	```go
	value, ok := b.GetFloat32("price") // read typed. !ok -> field is unset or no such field in the schema. Works faster and takes less memory allocations than Get()
	b.Get("price") // read untyped. nil -> field is unset or no such field in the schema
	b.Set("price", nil) // set to nil means unset
	bytes = b.ToBytes()
	```
- See [dynobuffers_test.go](dynobuffers_test.go) for usage examples

# Benchmarks
## Description
- [benchmarks\benchRead_test.go](benchmarks\benchRead_test.go) read benchmarks comparing to Avro, FlatBuffers, JSON
- [benchmarks\benchReadWrite_test.go](benchmarks\benchReadWrite_test.go) read and write benchmarks comparing to Avro, FlatBuffers, JSON
- Benchmarks naming
  - BenchmarkWrite... - benchmark read, change 1 field and write
  - BenchmarkRead... - benchmark read only
  - ...Simple... - schema with 3 fields, read and multiply 2 fields
  - ...Article... - schema with 123 fields
  - ...ReadFewFields... - read and multiply 2 fields
  - ...ReadAllFields... - read all fields
## Results
- Schema of 123 fields, read and multiply 2 fields, change 1 field and write
```
DynoBuffers    	  	    200000	      8763 ns/op	    5240 B/op	      41 allocs/op
FlatBuffers         	    300000	      5702 ns/op	    3672 B/op	      19 allocs/op
Json                	     10000	    201575 ns/op	   29096 B/op	     857 allocs/op
LinkedInAvro                100000	     25173 ns/op	   11929 B/op	     176 allocs/op
```
- Schema of 3 fields, read and multiply 2 fields
```
DynoBuffers               20000000	      65.0 ns/op	       0 B/op	       0 allocs/op
FlatBuffers               50000000	      33.2 ns/op	       0 B/op	       0 allocs/op
Json                        500000	      2660 ns/op	     760 B/op	      21 allocs/op
LinkedInAvro               2000000	       859 ns/op	     744 B/op	      10 allocs/op
```
- Schema of 123 fields, read and multiple 2 fields
```
DynoBuffers     	  20000000	      83.2 ns/op	       0 B/op	       0 allocs/op
FlatBuffers          	 100000000	      26.5 ns/op	       0 B/op	       0 allocs/op
Json                 	     10000	    108663 ns/op	   13730 B/op	     604 allocs/op
LinkedInAvro         	    100000	     14936 ns/op	   11257 B/op	     149 allocs/op
```
- Schema of 123 fields, simply read all fields
```
DynoBuffers      	    300000	      4935 ns/op	      35 B/op	      11 allocs/op
FlatBuffers          	   1000000	      1421 ns/op	       0 B/op	       0 allocs/op
Json                 	     10000	    116814 ns/op	   14130 B/op	     603 allocs/op
LinkedInAvro         	    100000	     19187 ns/op	   11257 B/op	     149 allocs/op
```

# To do
- Lists?
- Written in New -> read in Old -> write in Old -> New fields are kept.