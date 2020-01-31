module github.com/untillpro/dynobuffers

go 1.13

require (
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/flatbuffers v1.11.0
	github.com/kr/pretty v0.1.0 // indirect
	github.com/kylelemons/godebug v1.1.0
	github.com/linkedin/goavro v2.1.0+incompatible
	github.com/stretchr/testify v1.4.0
	github.com/untillpro/airs-iqueues v0.0.0-20200117121023-103366887fc7 // indirect
	github.com/untillpro/godif v0.16.0 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/linkedin/goavro.v1 v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/untillpro/dynobuffers => ../ // for benchmarks
