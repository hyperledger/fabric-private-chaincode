#!/usr/bin/env bash
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Checks what FPC_CGO switches: without it, make builds and tests only the Go
# side and refuses to build the ecc binary or anything in SGX_MODE=HW; with it
# (as in the dev container), the C side is built as before.
set -euo pipefail
cd "$(dirname "$0")/.."
# run as a developer would on the host, regardless of the calling environment
unset FPC_CGO

fail() { echo "FAIL: $*" >&2; exit 1; }

# 1. without FPC_CGO, make never enters the C side
c_dirs='(protos|sgx|cmake)( |$|/)'
for t in build test; do
    out="$(make -n "$t" 2>&1 || true)"
    if grep -qE "\-C ${c_dirs}" <<<"${out}"; then
        fail "'make ${t}' without FPC_CGO enters the C side: $(grep -oE "\-C ${c_dirs}" <<<"${out}" | sort -u | tr '\n' ' ')"
    fi
done

# 2. with FPC_CGO=1, it does, in today's order
# (read from make's database: macOS ships make 3.81, which has no --eval)
order="$(FPC_CGO=1 make -pRrq -f Makefile : 2>/dev/null | sed -n 's/^SUB_DIRS = //p' | sed 's/ *#.*//; s/ *$//' || true)"
[ "${order}" = "protos sgx/common internal ercc sgx/ecc_enclave ecc fabric client_sdk samples sgx/samples utils integration sgx/integration" ] ||
    fail "FPC_CGO=1 SUB_DIRS changed: ${order}"

# 3. no ecc binary without the C side
if out="$(make -s -C ecc ecc 2>&1)"; then fail "make -C ecc ecc succeeded without FPC_CGO"; fi
grep -q "needs the C enclave; run with FPC_CGO=1" <<<"${out}" || fail "unexpected ecc message: ${out}"

# 4. no HW mode without the C side
if out="$(SGX_MODE=HW make -s -C ercc ercc 2>&1)"; then fail "SGX_MODE=HW make -C ercc ercc succeeded without FPC_CGO"; fi
grep -q "SGX_MODE=HW requires FPC_CGO=1" <<<"${out}" || fail "unexpected ercc message: ${out}"

# 5. HW mode is still started from the host through the dev container
out="$(SGX_MODE=HW make -n -C utils/docker run-dev DOCKER_DEV_SKIP_BUILD=1 DOCKER_DEV_OPTIONAL_CMD=true 2>&1)" ||
    fail "SGX_MODE=HW make -C utils/docker run-dev fails: ${out}"
grep -q -- "--env FPC_CGO=1" <<<"${out}" || fail "run-dev does not pass FPC_CGO=1"
grep -q -- "--env SGX_MODE=HW" <<<"${out}" || fail "run-dev does not pass SGX_MODE=HW"

# 6. the suites (all need the dev container for now) only with FPC_CGO=1
[ -z "$(make -s suites)" ] || fail "make suites without FPC_CGO lists: $(make -s suites)"
[ -n "$(FPC_CGO=1 make -s suites)" ] || fail "FPC_CGO=1 make suites lists nothing"

echo "FPC_CGO OK"
