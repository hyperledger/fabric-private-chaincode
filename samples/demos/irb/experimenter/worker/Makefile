# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

PID_FILE=.worker.pid

PID="$(shell cat ${PID_FILE} 2> /dev/null)"

run:
	@if [ -z ${PID} ]; then \
	    PYTHONPATH=../../ python3 workerCLI.py &  echo -n "$$!" > ${PID_FILE} ;\
	fi

stop:
	@if [ ! -z ${PID} ]; then \
	    wget --quiet --spider localhost:5000/shutdown -O /dev/null ;\
	    rm -rf ${PID_FILE} ;\
	fi

