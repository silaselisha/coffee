
grpc:
	sudo rm -rf pkg/pb/*.go
	protoc --proto_path=pkg/proto --go_out=pkg/pb --go_opt=paths=source_relative \
	--grpc-gateway_out=pkg/pb \
	 --grpc-gateway_opt=paths=source_relative \
	--go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative \
	pkg/proto/*.proto

.PHONY: grpc 