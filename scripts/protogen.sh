#!/bin/bash

PROTOS_PATH="$(dirname "$(dirname "$(readlink "$0")")")"

protoc -I ${PROTOS_PATH}/ \
  --go_out=plugins=grpc:${PROTOS_PATH} \
  ${PROTOS_PATH}/protos/tendermint/tendermint.proto