#!/bin/bash
#
# Copyright Intel Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

export DEMO_CLIENT_SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
export FPC_PATH="${FPC_PATH:-${DEMO_CLIENT_SCRIPTS_DIR}/../../..}"

START_CMD="${FPC_PATH}"/demo/scripts/startFPCAuctionNetwork.sh
TEARDOWN_CMD="${FPC_PATH}"/demo/scripts/teardown.sh


. "${DEMO_CLIENT_SCRIPTS_DIR}"/lib/dsl.sh


help() {
    echo "$(basename $0) [--help|-h|-?] [--bootstrap|-b] [--dry-run|-d] [--non-interactive|-n] [--skip-delay|-s] [--mock-reset|-r] <script-file>
    Run the demo scenario codified in the passed script file.
    - If you pass option --bootstrap, it will also first bring up an FPC network
      and tear it down at the end; otherwise it assumes you have already
      a running setup ...
    - option --dry-run/-d will just display all requests but not execute/submit
      any of them
    - option --non-interactive/-n will case _all_ requests to be submited even
      requests from submit_manual.  This allows you to easily validate all
      json files, even when some steps would be manual in an actual scenario
    - option --skip-delay/-s allows you ignore all delays to speed-up demo
    - option --mock-reset/-r will try reset the mock-server at the beginning
      (obviously, this won't work if you run against the fabric-gateway backend;
       to achieve the equivalent for fabric-gateway, use option --bootstrap)
"
}

# argument handling
# - defaults
bootstrap=false
mock_reset=false
while [[ $# > 0 ]] && [[ $1 =~ "-" ]];
do
    opt=$1
    case $opt in
        --bootstrap|-b)
	    bootstrap=true
	    ;;

        --skip-delay|-s)
	    skip_delay=true
	    ;;

	--dry-run|-d)
	    dry_run=true
	    ;;

	--non-interactive|-n)
	    non_interactive=true
	    ;;

	--mock-reset|-r)
	    mock_reset=true
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
if ${bootstrap} && ${mock_reset}; then
    echo "ERROR: bootstrap and mock-reset options are mutually exclusive"
    help
    exit 1
fi

# (optional) which also sets up the whole fabric network
if ${bootstrap}; then
    $START_CMD --build-cc || die "could not bring up fpc network"
    sleep 10 # give some time for the client infrastructure to start ...
fi

if ${mock_reset}; then
    curl -X GET "http://localhost:3000/api/demo/start" || die "could not reset mock server"
fi

# execute the script
. "${scriptFile}" || die "executing script '${scriptFile}' failed ..."

if ${bootstrap}; then
    $TEARDOWN_CMD || die "could not bring down fpc network"
fi
