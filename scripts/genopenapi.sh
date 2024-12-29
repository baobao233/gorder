#!/usr/bin/env bash

set -euo pipefail
# enables globstar, using `**`.

shopt -s globstar

if ! [[ "$0" =~ scripts/genopenapi.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

source ./scripts/lib.sh

OPENAPI_ROOT="./api/openapi"

# 为框架服务生成代码，只能有一个元素，可以具体看 openapi 的 github 文档
GEN_SERVER=(
  # "chi-server"
  "gin-server"
)
  
# 检查是否只有一个元素
if [ "${#GEN_SERVER[@]}" -ne 1 ]; then
    log_error "GEN_SERVER enables more than 1 server, please check"
    exit 255
fi

# 打印当前的 server
log_callout "Using ${GEN_SERVER[0]}"

# 找到 proto 文件
function openapi_files {
    openapi_files=$(ls ${OPENAPI_ROOT})
    echo "${openapi_files[@]}"
}

# 参数：output_dir，package_name，service
function gen() {
  local output_dir=$1
  local package=$2
  local service=$3

  # 创建文件夹
  run mkdir -p "$output_dir"
  # 如果文件夹中有东西去要清除再重新生成
  run find "$output_dir" -type f -name "*.gen.go" -delete

  # 准备客户端代码的目录，脚本在 lib.sh
  prepare_dir "internal/common/client/$service"

  # 服务端代码
  run oapi-codegen -config api/openapi/cfg.yaml -generate types -o "$output_dir/openapi_types.gen.go" -package "$package" "api/openapi/$service.yml"
  run oapi-codegen -config api/openapi/cfg.yaml -generate "$GEN_SERVER" -o "$output_dir/openapi_api.gen.go" -package "$package" "api/openapi/$service.yml"

  # 客户端代码，type 其实就是struct 数据结构
  run oapi-codegen -config api/openapi/cfg.yaml -generate client -o "internal/common/client/$service/openapi_client.gen.go" -package "$service" "api/openapi/$service.yml"
  run oapi-codegen -config api/openapi/cfg.yaml -generate types -o "internal/common/client/$service/openapi_types.gen.go" -package "$service" "api/openapi/$service.yml"
}

gen internal/order/ports ports order

log_success "openapi generate success"