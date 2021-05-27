# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

import os
import json
import base64

def GetAttestation():
    sgx_mode = os.environ.get("SGX_MODE", "SIM")

    print("Get attestation in SGX {0} mode".format(sgx_mode))

    if sgx_mode == "SIM":
        attestation = dict()
        attestation['attestation_type'] = "simulated"
        attestation['attestation'] = base64.b64encode(b"0").decode()
        return json.dumps(attestation, separators=(',', ':'))

    if sgx_mode == "HW":
        return str("TODO: get Graphene attestation")
