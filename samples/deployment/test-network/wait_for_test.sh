#!/usr/bin/env bash
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Tests for wait_for.sh. Run: ./wait_for_test.sh
set -euo pipefail
cd "$(dirname "$0")"
. ./wait_for.sh

fail() { echo "FAIL: $*" >&2; exit 1; }

# succeeds immediately
wait_for 4 "true" true || fail "wait_for should succeed when the command succeeds"

# times out, returns 1, names what it waited for
start=$SECONDS
if err="$(wait_for 4 "the thing" false 2>&1)"; then
    fail "wait_for should fail when the command never succeeds"
fi
[ $((SECONDS - start)) -le 6 ] || fail "wait_for took longer than its timeout"
grep -q "timeout after 4s waiting for the thing" <<<"${err}" || fail "unexpected message: ${err}"

# succeeds once the command starts succeeding
marker="$(mktemp)"
rm -f "${marker}"
second_call() { if [ -f "${marker}" ]; then return 0; fi; touch "${marker}"; return 1; }
wait_for 6 "second call" second_call || fail "wait_for should succeed on a later attempt"
rm -f "${marker}"

echo "wait_for OK"
