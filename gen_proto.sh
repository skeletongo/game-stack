#!/usr/bin/env bash
# 生成所有模块的 proto 代码（player.pb.go / player_grpc.pb.go 等）。
# proto/*.proto → 生成到 proto/<name>/ 子目录（与 .proto 源文件分离）
# module/*/grpc/*.proto → 生成到同目录（与 .proto 同级）
# 使用方法：bash gen_proto.sh
set -euo pipefail

echo "=== generating proto files ==="

for proto in proto/*.proto; do
    echo "  $proto"
    protoc --proto_path=. \
        --go_out=. --go_opt=module=github.com/skeletongo/game-stack \
        --go-grpc_out=. --go-grpc_opt=module=github.com/skeletongo/game-stack \
        "$proto"
done

for proto in module/*/*.proto; do
    echo "  $proto"
    protoc --proto_path=. \
        --go_out=. --go_opt=module=github.com/skeletongo/game-stack \
        --go-grpc_out=. --go-grpc_opt=module=github.com/skeletongo/game-stack \
        "$proto"
done

echo "=== done ==="
