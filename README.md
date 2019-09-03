# Dyno Buffers

Fielded byte array with get\set by name

# Abilities
- 
- Read raw bytes from array converting to the required type described in the Schema
- No codegen, no wrappers, no compilers. Just fields description and get\set by name
- Non-set, unset and set to nil fields takes no place
- Supported types
  - `int32, int64, float32, float64, bool, string (var-size), byte`
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
	err := b.Set("price", float32(0.123))
	if err != nil {
		panic(err) // current schema has no such field or value of wrong type is provided
	}
	b.Set("name", nil)
	```
- To bytes array
	```go
	bytes := b.ToBytes()
	```
- Read Buffer from bytes using Schema
	```go
	b = dynobuffers.ReadBuffer(bytes, schema)
	```
- Work with Buffer 
	```go
	value, isSet := b.Get("price")
	b.Set("price", nil)
	b.Unset("name")
	bytes = b.ToBytes()
	```
- See [dynobuffers_test.go](dynobuffers_test.go) for usage examples

# Binary format
All offsets are non-relative. All fields are ordered by the Schema

| var-size values offsets pos | fixed-size values pos | bit mask | fixed-size values | var-size values offsets | var-size values |
| :---: | :---: | :---: | :---: | :---: | :---: |
| 4 bytes | 4 bytes | len (fieldsAmount/4 + 1) | len sum(fixed-size fields sizes) | len (var-size fields amount)*8 | |
| |	| for each field in the buffer: 1st bit - is field set, 2nd bit - is field set to nil | field types are taken from the Schema  | for each field: 1nd 4 bytes - value pos, 2nd 4 bytes - value size                 |                 |
| | |           |       |            |                 |
															

# To do
- Lists of nested objects
- Lists of primitives?
- Written in New -> read in Old -> write in Old -> New fields are kept.
- Store numbers in var-size Avro format?
