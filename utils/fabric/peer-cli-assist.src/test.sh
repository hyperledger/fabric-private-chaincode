#!/bin/bash
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

#set -ex 

result_pipe=$(mktemp -u -t testpipeXXXX)
mkfifo $result_pipe

cc_ek="a-key"
exec  3> >( ../peer-cli-assist handleRequestAndResponse "${cc_ek}" "${result_pipe}")
assist_pid=$!

exec 4<${result_pipe}

echo >&3 '{"Function":"init", "Args": ["MyAuctionHouse"]}'

read <&4 encrypted_request
echo "encrypted_request='$encrypted_request'"

echo >&3 "I'm supposed to be an base64 encoded serialized ChaincodeResponseMessage"
read <&4 decrypted_response
echo "decrypted_response='$decrypted_response'"

wait ${assist_pid}
echo "child exit code=$?"

rm $result_pipe
