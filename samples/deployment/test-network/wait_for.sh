# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# wait_for <timeout_seconds> <description> <command...>
#
# Runs the command every 2 seconds until it succeeds. Returns 1 and prints
# "timeout after <timeout_seconds>s waiting for <description>" if it never does.
wait_for() {
  local timeout="$1" description="$2"
  shift 2
  local deadline=$((SECONDS + timeout))
  until "$@" >/dev/null 2>&1; do
    if [ "${SECONDS}" -ge "${deadline}" ]; then
      echo "timeout after ${timeout}s waiting for ${description}" >&2
      return 1
    fi
    sleep 2
  done
}
