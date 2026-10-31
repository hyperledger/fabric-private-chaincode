#!/usr/bin/env bash
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Checks that the per-suite make targets are consistent: every suite listed by
# `make suites` has a target, and `make test` runs all of them.
set -euo pipefail
cd "$(dirname "$0")/.."

expected="${EXPECTED_SUITES:-auction echo kv deployment client_sdk go_echo go_auction go_kv go_simulation crash irb test_network}"
actual="$(make -s suites)"
if [ "${actual}" != "${expected}" ]; then
    echo "make suites: expected '${expected}', got '${actual}'" >&2
    exit 1
fi

# every suite needs its own rule with a recipe (the .PHONY line alone would
# make it a known target, so check the Makefile text, not make's database)
for s in ${actual}; do
    if ! grep -A1 -E "^suite-${s}:" Makefile | grep -q $'^\t'; then
        echo "missing make target with a recipe: suite-${s}" >&2
        exit 1
    fi
done

# the recipe of `test` itself
test_plan="$(sed -n '/^test:/,/^$/p' Makefile)"
if ! grep -q 'unit-test' <<<"${test_plan}" || ! grep -q 'suite-$(SUITE)' <<<"${test_plan}"; then
    echo "make test does not run unit-test and every suite-\$(SUITE)" >&2
    exit 1
fi
echo "suites OK: ${actual}"
