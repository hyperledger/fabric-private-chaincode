#!/usr/bin/env bash
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Checks which enclave the ecc binary uses: the mock (ecc/chaincode/enclave)
# without fpc_cgo, the real cgo enclave (sgx/ecc/enclave) with it. With
# FPC_CGO=1 (dev container, after the C side is built) it also checks that the
# built ecc binary links the enclave library.
set -euo pipefail
cd "$(dirname "$0")/.."

fail() { echo "FAIL: $*" >&2; exit 1; }
mod=github.com/hyperledger/fabric-private-chaincode
real="${mod}/sgx/ecc/enclave"
mock="${mod}/ecc/chaincode/enclave"

deps="$(go list -deps ./ecc)"
grep -qx "${mock}" <<<"${deps}" || fail "ecc without fpc_cgo does not use the mock enclave"
if grep -qx "${real}" <<<"${deps}"; then fail "ecc without fpc_cgo uses the real enclave"; fi

deps="$(go list -deps -tags fpc_cgo ./ecc)"
grep -qx "${real}" <<<"${deps}" || fail "ecc with fpc_cgo does not use the real enclave"
if grep -qx "${mock}" <<<"${deps}"; then fail "ecc with fpc_cgo uses the mock enclave"; fi

# ecc/Makefile must build the package, not main.go alone: naming files makes go
# skip the tagged enclave_cgo.go/enclave_mock.go, so newEnclave is undefined
cmd="$(make -s -n -C ecc ecc FPC_CGO=1 2>/dev/null | grep -E '^go +build' || true)"
[ -n "${cmd}" ] || fail "no go build in 'make -C ecc ecc'"
if grep -qE '\.go( |$)' <<<"${cmd}"; then fail "ecc is built from a file list, not the package: ${cmd}"; fi
(cd ecc && go build -o /dev/null .) || fail "ecc does not build as a package"

if [ "${FPC_CGO:-}" = 1 ] && [ -x ecc/ecc ]; then
    ldd ecc/ecc | grep -q libsgxcc || fail "ecc/ecc does not link libsgxcc"
fi
echo "ecc enclave OK"
