module github.com/DataDog/datadog-agent/internal/tools/proto

go 1.25.0

require (
	github.com/favadi/protoc-go-inject-tag v1.4.0
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.5.4
	github.com/grpc-ecosystem/grpc-gateway v1.12.2
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10
	github.com/tinylib/msgp v1.1.6
	google.golang.org/grpc v1.82.1
)

require (
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/golang/glog v1.2.5 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.43.0 // indirect
	google.golang.org/genproto v0.0.0-20190927181202-20e1ac93f88c // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace google.golang.org/protobuf v1.33.0 => google.golang.org/protobuf v1.34.0
