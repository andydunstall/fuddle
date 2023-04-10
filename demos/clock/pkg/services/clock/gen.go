package clock

//go:generate protoc -I . --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. rpc.proto
