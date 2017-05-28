//
// Generate Protobuf services and messages
//
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./protos:. --gofast_out=plugins=grpc:./services/ ./protos/certificates/certificates.proto
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./protos:. --gofast_out=plugins=grpc:./services/ ./protos/nodes/nodes.proto
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./protos:. --gofast_out=plugins=grpc:./services/ ./protos/users/users.proto
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./protos:. --gofast_out=plugins=grpc:./services/ ./protos/sites/sites.proto
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./protos:. --gofast_out=plugins=grpc:./services/ ./protos/rules/rules.proto

package waffy
