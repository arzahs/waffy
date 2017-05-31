//
// Generate Protobuf services and messages
//
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./pkg/services/protos:. --gofast_out=plugins=grpc:./ ./pkg/services/protos/certificates/certificates.proto
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./pkg/services/protos:. --gofast_out=plugins=grpc:./ ./pkg/services/protos/nodes/nodes.proto
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./pkg/services/protos:. --gofast_out=plugins=grpc:./ ./pkg/services/protos/users/users.proto
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./pkg/services/protos:. --gofast_out=plugins=grpc:./ ./pkg/services/protos/sites/sites.proto
//go:generate protoc --proto_path=.:../../../:./vendor:./vendor/github.com/gogo/protobuf/protobuf:./pkg/services/protos:. --gofast_out=plugins=grpc:./ ./pkg/services/protos/rules/rules.proto

package waffy
