#!/bin/sh
#
# Copyright IBM Corp, SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
# to update CHANGELOG.md, run this script with the latest commit of the latest release number found in CHANGELOG.md
#
#
# Usage: ./script/changelog.sh commit-of-pre-release new-release-commit release-name

echo "## $3\n$(date)" >> CHANGELOG.new
echo "" >> CHANGELOG.new
git log $1..$2 --oneline | grep -v Merge | sed -e "s/\([0-9|a-z]*\)/* \[\1\](https:\/\/github.com\/hyperledger\/fabric-private-chaincode\/commit\/\1)/" >> CHANGELOG.new
echo "" >> CHANGELOG.new
cat CHANGELOG.md >> CHANGELOG.new
mv -f CHANGELOG.new CHANGELOG.md
