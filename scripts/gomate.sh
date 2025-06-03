#!/usr/bin/env bash
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -euo pipefail

# lint me: shfmt -i 2 -ci -l -w

IFS=$'\t\n' # Split on newlines and tabs (but not on spaces)
script_name=$(basename "${0}")
script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
readonly script_name script_dir

readonly repo_dir="$(pwd)"

# how to use this script
function usage() {
  cli_name=${0##*/}
  echo "gomate helps to manage multiple go modules in a single repository.
Usage: $cli_name [command]
Commands:
  initwork      creates go workspace for this project (\`go work init\` and \`go work use\` everywhere)
  tidy          runs \`go mod tidy\` everywhere
  update [XYZ]  updates a specific dep (\`go get XYZ\`) everywhere; if no dep argument given, \`go get -u\` is called
  help          shows this help
  "
  exit 1
}

# create go work init
function init_work() {
  echo "go work init"
  go work init
  find "$repo_dir" -iname "go.mod" -exec dirname {} \; | xargs go work use
}

# update deps; take as parameter the dependency to update; if empty all deps are updates
function update() {
  if [[ -z $1 ]]; then
    # check all update
    echo "go get -u everywhere"
    find "$repo_dir" -iname "go.mod" -execdir sh -c "go get -u" \;
  else
    # check a specific dep
    echo "go get $1 everywhere"
    find "$repo_dir" -iname "go.mod" -execdir sh -c "go get $1" \;
  fi
}

# run go mod tidy everywhere
function tidy() {
  echo "go mod tidy everywhere"
  find "$repo_dir" -iname "go.mod" -execdir sh -c "go mod tidy" \;
}

main() {
  case "${1-""}" in
    initwork)
      init_work
      ;;
    tidy)
      tidy
      ;;
    update)
      update "${2-""}"
      ;;
    *)
      usage
      ;;
  esac
}

main "${@}"
