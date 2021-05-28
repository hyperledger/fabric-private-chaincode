#!/bin/bash

# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

set -e

BASE=..

terminate() {
    echo "Ledger: stop"
    (cd ${BASE}/ledger-helper && make stop)

    echo "Storage Service: stop"
    (cd ${BASE}/storage-service && make stop)

    echo "Experiment worker: stop"
    (cd ${BASE}/experimenter/worker && make stop-docker)
}

trap terminate EXIT

echo "BEGIN IRB Test"

echo "Storage Service: start"
(cd ${BASE}/storage-service && make run)

echo "Experiment worker: start"
(cd ${BASE}/experimenter/worker && make run-docker)

echo "Ledger: start"
(cd ${BASE}/ledger-helper && make stop && make run)

echo "Data provider: create users, data and register data"
(cd ${BASE}/data-provider/test && ./test)

echo "Principal investigator: register study"
(cd ${BASE}/principal-investigator/test && ./test registerstudy)

echo "Experimenter: get worker credential and register experiment"
(cd ${BASE}/experimenter/client/test && ./test newexperiment)

echo "Principal investigator: approve experiment"
(cd ${BASE}/principal-investigator/test && ./test approveexperiment)

echo "Experimenter: get evalution pack and execute it"
(cd ${BASE}/experimenter/client/test && ./test executeevaluationpack)

echo "END IRB Test"
