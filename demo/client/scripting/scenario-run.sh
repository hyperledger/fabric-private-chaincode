#!/bin/bash
#
# Copyright Intel Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

export DEMO_CLIENT_SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
export FPC_ROOT="${DEMO_CLIENT_SCRIPTS_DIR}"/../../..

START_CMD="${FPC_ROOT}"/demo/scripts/startFPCAuctionNetwork.sh
TEARDOWN_CMD="${FPC_ROOT}"/demo/scripts/teardown.sh


. "${DEMO_CLIENT_SCRIPTS_DIR}"/lib/dsl.sh


help() {
    echo "$(basename $0) [--help|-h|-?] [--bootstrap|-b] [--dry-run|-d] [--non-interactive|-n] <script-file>
    Run the demo scenario codified in the passed script file.
    - If you pass option --bootstrap, it will also first bring up an FPC network
      and tear it down at the end; otherwise it assumes you have already
      a running setup ...
    - option --dry-run/-d will just display all requests but not execute/submit
      any of them
    - option --non-interactive/-n will case _all_ requests to be submited even
      requests from submit_manual.  This allows you to easily validate all
      json files, even when some steps would be manual in an actual scenario
"
}

# argument handling
# - defaults
typeset -i bootstrap=0
while [[ $# > 0 ]] && [[ $1 =~ "-" ]];
do
    opt=$1
    case $opt in
        --bootstrap|-b)
	    bootstrap=1
	    ;;

        --bootstrap|-b)
	    bootstrap=1
	    ;;

	--dry-run|-d)
	    dry_run=1
	    ;;

	--non-interactive|-n)
	    non_interactive=1
	    ;;

        --help|-h|-\?)
	    help
	    exit 0
	    ;;
	*)
            echo "ERROR: unknown option $opt."
	    help
	    exit 1
    esac
    shift # past argument or value
done
if [ $# -ne 1 ]; then
            echo "ERROR: missing script file."
	    help
	    exit 1
fi
scriptFile=$1
if [ ! -f $scriptFile  ]; then
            echo "ERROR: missing script file $scriptFile"
	    help
	    exit 1
fi


# (optional) which also sets up the whole fabric network
if [ ${bootstrap} -eq 1 ]; then
    # START_CMD does _not_ build the FPC peer and it is expensive to build it, 
    # so let's look first whether we have it ...
    if [ -z "$(docker images | grep hyperledger/fabric-peer-fpc )" ]; then
        # .. and if not, build it
    	pushd "${FPC_ROOT}/utils/docker" || die "can't go to peer build location"
        make peer || die "can't build peer"
    fi
    # START_CMD also does not call utils/docker-compose/scripts/bootstrap!
    # as docker(-compose) will download images, we care only about binaries
    # but download only if we do not have already executables installed and/or
    # in path. We also assume that if you have cryptogen in path, you probably
    # have the other needed, e.g., configtxgen, as well ...
    if [ ! -d "${FPC_ROOT}/utils/docker-compose/bin" ] && [ -z "$(which cryptogen)" ]; then 
        "${FPC_ROOT}/utils/docker-compose/scripts/bootstrap.sh" -d
    fi
    # Note: Above could eventually be moved to START_CMD once the various PRs
    # are merged ...
    $START_CMD || die "could not bring up fpc network"
    sleep 10 # give some time for the client infrastructure to start ...
fi

# execute the script
. "${scriptFile}"

if [ ${bootstrap} -eq 1 ]; then
    $TEARDOWN_CMD || die "could not bring down fpc network"
fi
