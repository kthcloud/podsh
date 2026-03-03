//go:generate mkdir -p api
//go:generate sh -c "PATH=$PATH:./wrappers protoc --proto_path=. --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative identity_events.proto"

package apiv1

