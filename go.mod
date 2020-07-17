module github.com/starjiang/easycall

go 1.14

require (
	github.com/aws/aws-sdk-go v1.33.6
	github.com/coreos/etcd v3.3.22+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/panjf2000/ants/v2 v2.4.1
	github.com/starjiang/elog v1.0.6
	github.com/vmihailenco/msgpack v4.0.4+incompatible
	go.uber.org/zap v1.15.0 // indirect
	google.golang.org/genproto v0.0.0-20200709005830-7a2ca40e9dc3 // indirect
	google.golang.org/grpc v1.30.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
