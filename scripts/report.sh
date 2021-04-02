#!/bin/bash

# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

# The scripts simply calls the codecov scripts to upload the coverage stats.
# So the scripts automatically detect all the coverage files in the repo
# (generated after executing binaries compiled/linked with the --coverage flag).
# The coverage stats should be visible on the codecov site (you might have to create an account)
# and also joint with the other checks on github.

curl -s https://codecov.io/bash | bash -s -- -Z -c ${CODECOV_REPO_TOKEN+"-t ${CODECOV_REPO_TOKEN}"}
if [ $? -ne 0 ]; then
    echo "Something went wrong while reporting coverage to codecov."
    echo "If you're running the codecov report locally,"
    echo "you might need to set the CODECOV_REPO_TOKEN environment variable, for example:"
    echo "export CODECOV_REPO_TOKEN=000000-00000-0000-000"
else
    echo "Coverage reported to Codecov"
fi
