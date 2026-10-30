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
  aligncheck    checks that all modules build each dependency with the same version
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

# check that every module of the repository builds each dependency with the same version;
# only dependencies whose packages are compiled count, as go mod tidy drops pins of any other module.
# The check covers the current platform; for another one, set GOOS and CGO_ENABLED=1 (cgo packages).
function align_check() {
  # each module resolves its own build list; a go.work would resolve all modules together
  export GOWORK=off
  # a replaced module counts with its replacement version; local path replaces are skipped
  local format='{{if .Error}}ERROR {{.ImportPath}}: {{.Error}}{{else}}{{with .Module}}{{if not .Main}}{{if .Replace}}{{if .Replace.Version}}{{.Path}} {{.Replace.Path}}@{{.Replace.Version}}{{end}}{{else}}{{.Path}} {{.Version}}{{end}}{{end}}{{end}}{{end}}'
  local versions="" mod dir out pkgs
  while read -r mod; do
    dir=$(dirname "$mod")
    pkgs=./...
    grep -q '^tool' "$repo_dir/$mod" && pkgs="./... tool"
    # tools and deps tags include the files that only keep dependencies in go.mod
    # shellcheck disable=SC2086 # pkgs is a list of package patterns
    if ! out=$(cd "$repo_dir/$dir" && IFS=' ' && go list -e -deps -test -tags tools,deps -f "$format" $pkgs); then
      echo "go list failed in $dir"
      return 1
    fi
    # a blank import of a main package in a tools.go file is the intended use of the tools tag
    if grep '^ERROR ' <<<"$out" | grep -v 'is a program, not an importable package'; then
      echo "cannot load the packages of $dir"
      return 1
    fi
    versions+=$(grep -v '^ERROR ' <<<"$out" | sort -u | sed "s|^|$dir |")$'\n'
  done < <(git -C "$repo_dir" ls-files '*go.mod')
  local mismatches
  mismatches=$(echo "$versions" | awk '{v[$2] = v[$2] " " $3 "(" $1 ")"; seen[$2 " " $3] = 1} END {for (k in seen) {split(k, a, " "); n[a[1]]++} for (p in n) if (n[p] > 1) print p ":" v[p]}' | sort)
  if [[ -n $mismatches ]]; then
    echo "dependencies built with different versions:"
    echo "$mismatches"
    return 1
  fi
  echo "all modules build their dependencies with the same versions"
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
    aligncheck)
      align_check
      ;;
    *)
      usage
      ;;
  esac
}

main "${@}"
