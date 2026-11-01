#!/usr/bin/env bash
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Checks how the build passes Go build tags: one opt-in tag (fpc_cgo) for the
# C side, always as a single -tags flag (go honours only the last -tags flag).
set -euo pipefail
cd "$(dirname "$0")/.."

fail() { echo "FAIL: $*" >&2; exit 1; }

# the go command lines `make -n` would run in a directory
go_cmds() { make -s -n -C "$1" "${@:2}" 2>/dev/null | grep -E '^[^#]*\bgo +(build|test|run|vet)\b' || true; }

# C side on, HW mode: one -tags flag with both tags
cmds="$(go_cmds internal test FPC_CGO=1 SGX_MODE=HW)"
[ -n "${cmds}" ] || fail "no go command in 'make -C internal test'"
while read -r line; do
    n="$(grep -o -- '-tags' <<<"${line}" | wc -l | tr -d ' ')"
    [ "${n}" = 1 ] || fail "expected one -tags flag, got ${n}: ${line}"
    grep -q -- '-tags [^ ]*fpc_cgo' <<<"${line}" || fail "fpc_cgo missing: ${line}"
    grep -q -- '-tags [^ ]*sgx_hw_mode' <<<"${line}" || fail "sgx_hw_mode missing: ${line}"
done <<<"${cmds}"

# default: pure Go, no tags at all
cmds="$(go_cmds internal test)"
if grep -q -- '-tags' <<<"${cmds}"; then fail "default build passes tags: ${cmds}"; fi

# ecc unit tests always use the mock enclave, even with the C side on
cmds="$(go_cmds ecc test FPC_CGO=1)"
[ -n "${cmds}" ] || fail "no go command in 'make -C ecc test'"
if grep -q fpc_cgo <<<"${cmds}"; then fail "ecc unit tests must not use fpc_cgo: ${cmds}"; fi

# FPC-go chaincode never gets fpc_cgo: ego would pull the PDO crypto (and
# OpenSSL) into the enclave, which then fails to load
cmds="$(make -s -n -C samples/chaincode/echo-go ecc FPC_CGO=1 2>/dev/null | grep -E '^ego-go +build' || true)"
[ -n "${cmds}" ] || fail "no ego-go build in 'make -C samples/chaincode/echo-go ecc'"
if grep -q fpc_cgo <<<"${cmds}"; then fail "FPC-go chaincode must not use fpc_cgo: ${cmds}"; fi

# the old tags are gone
if git grep -n "mock_ecc\|WITH_PDO_CRYPTO" -- . ':!scripts/check-gotags.sh'; then
    fail "old build tags still in use"
fi
echo "go tags OK"
