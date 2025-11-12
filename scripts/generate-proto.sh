#!/bin/bash

# Script to generate Go code from protobuf definitions

set -e

echo "Generating protobuf code..."

# Install protoc plugins if not present
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Installing protoc-gen-go-grpc..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Generate code from proto files
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/task_queue.proto

echo "Protobuf code generation complete!"
echo "Generated files:"
echo "  - proto/task_queue.pb.go"
echo "  - proto/task_queue_grpc.pb.go"
