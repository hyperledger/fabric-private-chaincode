#!/bin/bash

# Copyright Greg Haskins All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

# Note: This file is adapted from hyperledger/fabric scripts/golinter.sh (ec81f3e74)

set -e

if [ $# -eq 0 ] || [ -z $1 ]; then
    echo "Missing top folder input parameter for $0"
    exit 1
fi

# shellcheck source=/dev/null
source "$(cd "$(dirname "$0")" && pwd)/functions.sh"

echo "Checking with gofmt"
OUTPUT="$(gofmt -l -s $1)"
OUTPUT="$(filterExcludedAndGeneratedFiles "$OUTPUT")"
if [ -n "$OUTPUT" ]; then
    echo "The following files contain gofmt errors"
    echo "$OUTPUT"
    echo "The gofmt command 'gofmt -l -s -w' must be run for these files"
    exit 1
fi

echo "Checking with goimports"
OUTPUT="$(goimports -l $1)"
OUTPUT="$(filterExcludedAndGeneratedFiles "$OUTPUT")"
if [ -n "$OUTPUT" ]; then
    echo "The following files contain goimports errors"
    echo "$OUTPUT"
    echo "The goimports command 'goimports -l -w' must be run for these files"
    exit 1
fi

# Now that context is part of the standard library, we should use it
# consistently. The only place where the legacy golang.org version should be
# referenced is in the generated protos.
echo "Checking for golang.org/x/net/context"
# shellcheck disable=SC2016
TEMPLATE='{{with $d := .}}{{range $d.Imports}}{{ printf "%s:%s " $d.ImportPath . }}{{end}}{{end}}'
OUTPUT="$(go list -f "$TEMPLATE" $1/... | grep 'golang.org/x/net/context' | cut -f1 -d:)"
if [ -n "$OUTPUT" ]; then
    echo "The following packages import golang.org/x/net/context instead of context"
    echo "$OUTPUT"
    exit 1
fi

# We use golang/protobuf but goimports likes to add gogo/protobuf.
# Prevent accidental import of gogo/protobuf.
echo "Checking for github.com/gogo/protobuf"
# shellcheck disable=SC2016
TEMPLATE='{{with $d := .}}{{range $d.Imports}}{{ printf "%s:%s " $d.ImportPath . }}{{end}}{{end}}'
OUTPUT="$(go list -f "$TEMPLATE" $1/... | grep 'github.com/gogo/protobuf' | cut -f1 -d:)"
if [ -n "$OUTPUT" ]; then
    echo "The following packages import github.com/gogo/protobuf instead of github.com/golang/protobuf"
    echo "$OUTPUT"
    exit 1
fi

echo "Checking with go vet"
GOTAGS=WITH_PDO_CRYPTO
PRINTFUNCS="Debug,Debugf,Print,Printf,Info,Infof,Warning,Warningf,Error,Errorf,Critical,Criticalf,Sprint,Sprintf,Log,Logf,Panic,Panicf,Fatal,Fatalf,Notice,Noticef,Wrap,Wrapf,WithMessage"
OUTPUT="$(go vet -all -tags "$GOTAGS" -printfuncs "$PRINTFUNCS" $1/...)"
if [ -n "$OUTPUT" ]; then
    echo "The following files contain go vet errors"
    echo "$OUTPUT"
    exit 1
fi

echo "Checking with staticcheck"
OUTPUT="$(staticcheck -tags "$GOTAGS" $1/... || true)"
if [ -n "$OUTPUT" ]; then
    echo "The following staticcheck issues were flagged"
    echo "$OUTPUT"
    exit 1
fi

echo "Checking with misspell"
OUTPUT="$(misspell $(go list -f '{{.Dir}}' ./...))"
if [ -n "$OUTPUT" ]; then
    echo "The following files are have spelling errors:"
    echo "$OUTPUT"
    exit 1
fi

