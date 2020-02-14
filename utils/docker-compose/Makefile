# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

TOP = ../..
include $(TOP)/build.mk

build:
	# nothing to do, will all be done via docker-compose exercised by scripts ...

test:
	# we do nothing here. Note, though, that target test in demo/client/scripting
	# causes the scripts here to be exercised/tested ...

clobber:
	# first clean-up docker-compose network and demo related artifacts
	(cd ../../demo/; make clobber)
	# above presumably does below, but better safe than sorry :-)
	./scripts/teardown.sh --clean-slate
