#!/usr/bin/env bash
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Checks that the non-SGX suites (test_network, go_chaincode/simulation) do not
# overwrite the images and build outputs the SGX suites use, when suites run one
# after another as in `make test`. Runs inside the dev container after `make build`.
set -euo pipefail
cd "$(dirname "$0")/.."

image_id() { docker image inspect -f '{{.Id}}' "$1"; }

check_untouched() {
    local sample="$1" image="$2" before_cc before_ercc
    before_cc="$(image_id "${image}:latest")"
    before_ercc="$(image_id fpc/ercc:latest)"
    shift 2
    "$@"
    if [ "$(image_id "${image}:latest")" != "${before_cc}" ]; then
        echo "FAIL: '$*' overwrote ${image}" >&2
        exit 1
    fi
    if [ "$(image_id fpc/ercc:latest)" != "${before_ercc}" ]; then
        echo "FAIL: '$*' overwrote fpc/ercc" >&2
        exit 1
    fi
    if grep -q fake_mrenclave "samples/chaincode/${sample}/mrenclave"; then
        echo "FAIL: '$*' overwrote samples/chaincode/${sample}/mrenclave" >&2
        exit 1
    fi
}

make suite-go_echo
check_untouched echo-go fpc/echo-go make suite-test_network
image_id fpc/echo-go-nosgx:latest >/dev/null
image_id fpc/ercc-nosgx:latest >/dev/null

make suite-go_kv
check_untouched kv-test-go fpc/kv-test-go make -C integration/go_chaincode/simulation
image_id fpc/kv-test-no-sgx:latest >/dev/null

echo "suite images OK"
