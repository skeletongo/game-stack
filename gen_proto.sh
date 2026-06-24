#!/usr/bin/env bash
# 生成所有模块的 proto 代码（player.pb.go / player_grpc.pb.go 等）。
# proto/*.proto → 生成到 proto/<name>/ 子目录（与 .proto 源文件分离）
# module/*/rpc/*.proto → 按 go_package 生成到 module/<name>/rpc/grpc 等目录
# 使用方法：bash gen_proto.sh
set -euo pipefail

echo "=== generating proto files ==="

# 兼容 Windows 环境下通过 bash 调用 protoc.exe 的场景。
PROTOC_BIN="${PROTOC:-protoc}"
if ! command -v "$PROTOC_BIN" >/dev/null 2>&1; then
    if command -v protoc.exe >/dev/null 2>&1; then
        PROTOC_BIN="protoc.exe"
    else
        echo "protoc not found" >&2
        exit 1
    fi
fi

for proto in proto/*.proto; do
    echo "  $proto"
    "$PROTOC_BIN" --proto_path=. \
        --go_out=. --go_opt=module=github.com/skeletongo/game-stack \
        --go-grpc_out=. --go-grpc_opt=module=github.com/skeletongo/game-stack \
        "$proto"
done

# 模块 RPC 协议统一放在 module/<name>/rpc/，避免和模块根目录入口混杂。
for proto in module/*/rpc/*.proto; do
    echo "  $proto"
    "$PROTOC_BIN" --proto_path=. \
        --go_out=. --go_opt=module=github.com/skeletongo/game-stack \
        --go-grpc_out=. --go-grpc_opt=module=github.com/skeletongo/game-stack \
        "$proto"
done

echo "=== done ==="
