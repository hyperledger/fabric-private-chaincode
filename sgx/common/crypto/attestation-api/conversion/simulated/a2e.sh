#!/bin/bash

# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -e

###########################################################
# simulated_to_evidence
#   input:  evidence as parameter
#   output: SIMULATED_EVIDENCE variable
###########################################################
function simulated_to_evidence() {
    SIMULATED_EVIDENCE=$1
}
