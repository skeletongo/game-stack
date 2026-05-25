#!/usr/bin/env bash
# 生成所有模块的 proto 代码（player.pb.go / player_grpc.pb.go 等）。
# 生成文件与 .proto 放在同一目录。
# 使用方法：bash gen_proto.sh
set -euo pipefail

echo "=== generating proto files ==="

for proto in module/*/*.proto; do
    echo "  $proto"
    protoc --proto_path=. \
        --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        "$proto"
done

echo "=== done ==="
