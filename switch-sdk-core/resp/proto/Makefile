.PHONY: proto-gen clean-proto

# 生成 protobuf Go 代码
# 联动-proto->pb.go->response struct nedd pb.go
proto-gen:
	@echo "Generating protobuf Go code..."
	protoc --proto_path=. \
           --go_out=paths=source_relative:. \
           --go-grpc_out=paths=source_relative:. \
           *.proto
	@echo "Proto generation completed!"

# 清理生成的 proto 文件
clean-proto:
	@echo "Cleaning generated proto files..."
	rm -f *.pb.go
	@echo "Proto files cleaned!"

# 重新生成（先清理再生成）
proto-gen-clean: clean-proto proto-gen

# 帮助信息
help:
	@echo "Available commands:"
	@echo "  proto-gen        - Generate protobuf Go code"
	@echo "  clean-proto      - Clean generated proto files"
	@echo "  proto-gen-clean  - Clean and regenerate proto files"
	@echo "  help            - Show this help message"
