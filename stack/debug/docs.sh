#!/bin/bash
# 从 swag 注释生成 OpenAPI 文档。
# 在项目根目录执行：bash stack/debug/docs.sh
set -euo pipefail

DIR="$(cd "$(dirname "$0")" && pwd)"
swag init -g doc.go -d "$DIR" -o "$DIR/docs" --outputTypes json,yaml
echo "=== swagger docs generated at $DIR/docs ==="
