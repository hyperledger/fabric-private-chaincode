#!/usr/bin/env bash
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Checks the native (non-SGX) image build of a Go sample: `make` then
# `make docker` must package a static linux binary, also on a Mac, where the
# plain `make` builds a binary for the host.
set -euo pipefail
cd "$(dirname "$0")/.."
unset FPC_CGO

fail() { echo "FAIL: $*" >&2; exit 1; }
sample=samples/chaincode/echo-go
image=fpc/echo-go-native-check

make -s -C "${sample}" build
make -s -C "${sample}" docker DOCKER_IMAGE="${image}" DOCKER_QUIET_BUILD=1 >/dev/null

out="$(mktemp -d)"
trap 'rm -rf "${out}"; docker rmi -f "${image}:latest" "${image}:main" >/dev/null 2>&1 || true' EXIT
cid="$(docker create "${image}:latest")"
docker cp "${cid}:/usr/local/bin/chaincode" "${out}/chaincode" >/dev/null
docker rm "${cid}" >/dev/null

info="$(file "${out}/chaincode")"
grep -q "ELF" <<<"${info}" || fail "image binary is not a linux ELF: ${info}"
grep -q "statically linked" <<<"${info}" || fail "image binary is not static: ${info}"
echo "native image OK: ${info#*: }"
