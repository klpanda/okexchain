#!/usr/bin/env bash

set -eo pipefail
cosmos=vendor/github.com/cosmos/cosmos-sdk

proto_dir=$1

protoc -I "${cosmos}/proto" \
    -I "${cosmos}/third_party/proto" -I "${proto_dir}" \
    --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
    $(find "${proto_dir}" -name '*.proto')

cp -r github.com/okex/okchain/* ./
rm -rf github.com
