# Dyno Buffers

Fielded byte array with get\set by name

# Abilities
- Read raw bytes from array converting to the required type described in the Schema
- Non-set, unset and set to nil fields takes no place
- Supported types 
  - `int32, int64, float32, float64, bool, string`
- Schema versioning
  - Any data written with Schema of any version will be correctly read using Schema of any other version if new fields are appended to the end
    - Written in old Schema, read in New Schema -> nil result, field considered as unset
    - Written in new Schema, read in old Schema -> nil result, 

# Limitations
- Written in New -> read in Old -> write in Old -> New fields are lost (todo)
- No arrays (todo)
- No nested objects (todo)

# Usage
- Describe Schema
  - By yaml
  Field types:
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
	schema := NewSchema()
	schema.AddField("name", dynobuffers.FieldTypeString)
	schema.AddField("price", dynobuffers.FieldTypeFloat)
	schema.AddField("quantity", dynobuffers.FieldTypeInt)
	```
- Create Dyno Buffer using Schema
	```go
	b := NewBuffer(schema)
	```
- Set\modify fields according to the Schema
	```go
	err := b.Set("price", float32(0.123))
	if err != nil {
		panic(err) // value of wrong type is provided
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
- Work Buffer 
	```go
	value, isSet := b.Get("price") // value is interface{} of float32, isSet == true
	b.Set("price", nil)
	b.Unset("name")
	bytes = b.ToBytes()
	```
- See [dynobuffers_test.go](dynobuffers_test.go) for usage examples

#To do
- Lists of nested objects
- Lists of primitives?
- Written in New -> read in Old -> write in Old -> New fields are kept.
