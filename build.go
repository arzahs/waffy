//
// Generate Protobuf services and messages
//
//go:generate protoc --proto_path=.:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./protos:. --gofast_out=plugins=grpc:./services/ ./protos/certificates.proto
//go:generate protoc --proto_path=.:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./protos:. --gofast_out=plugins=grpc:./services/ ./protos/users.proto
//go:generate protoc --proto_path=.:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./protos:. --gofast_out=plugins=grpc:./services/ ./protos/sites.proto
//go:generate protoc --proto_path=.:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./protos:. --gofast_out=plugins=grpc:./services/ ./protos/rules.proto

package waffy
