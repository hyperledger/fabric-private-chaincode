#!/bin/bash
#
# Copyright Intel Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

# assumes FPC_PATH and DEMO_CLIENT_SCRIPTS_DIR is defined

. "${FPC_PATH}"/fabric/bin/lib/common_utils.sh


CLI="${DEMO_CLIENT_SCRIPTS_DIR}/cli"




#  Simple "DSL" to script auction scenarios
#---------------------------------------------
#
# TODO (maybe)
# - get auction status after submit, output when rounds/auction are finished
#   and/or do corresponding error handling?
#
# terms of language are
# - scenario [<scenario-path>] if path is missing, take directory where script file is located
# - "submit" <user> <action> [<expected return code, by default 0>]
# - "submit_manual" <user> <action>
# - "wait" <message> [<variable-for-user-input>]
# - "say" "some text"
# - "yell" "some important text"

# variables which influence the behaviour
dry_run=0		# never submit, just echo action
non_interactive=0	# always submit action, even submit_manual


# - state
typeset -i round=0

scenarioPath=""


# - commands
scenario() {
    scenarioPath="$1"

    if [[ -z "${scenarioPath}" ]]; then
	scenarioPath=$( dirname "${BASH_SOURCE[1]}") # 0 is this script and 1 should be caller ..
    fi
    say "Performing scenario '${scenarioPath}'"
    say "---------------------------------------"
}


submit() {
    user="$1"
    action="$2"
    expected_rc=$3

    if [ ${dry_run} -eq 1 ]; then
	say "dry-run mode: simulation action for submit '${user}' '${action}'"
	submit_raw 0 0 "${user}" "${action}" ${expected_rc}
    else
	submit_raw 1 0 "${user}" "${action}" ${expected_rc}
    fi
}

submit_manual() {
    user="$1"
    action="$2"

    if [ ${non_interactive} -eq 1 ]; then
	say "non-interactive mode: simulation action for submit '${user}' '${action}'"
	submit_raw 1 0 "${user}" "${action}" 
    else
	yell "create following request manually in the UI" # note, submit prints action & user, so no need to repeat here
	submit_raw 0 1 "${user}" "${action}" 
    fi
}

submit_raw() {
    do_it="$1"
    wait="$2"
    user="$3"
    action="$4"
    expected_rc=$5
    if [ -z "${expected_rc}" ]; then
	expected_rc=0;
    fi

    # scenario path
    # Note: two vars so we can quote scenarioPath properly ....
    if [[ ! -z "${scenarioPath}" ]]; then
	scenario_opt1="-scenario-path"
	scenario_opt2="${scenarioPath}"
    else
	scenario_opt1=""
	scenario_opt2=""
    fi

    # round handling
    # - iff startNextRound, increase round number ...
    if [[ "${action}" == "startNextRound" ]]; then
        #   Note: we assume this will succeed. Otherwise, we would have to test
        #   the return value below and reduce round again ...
	round=$round+1
    fi
    # iff action = submitClockBid, add round number
    if [[ "${action}" == "submitClockBid" ]]; then
	round_opt="-round $round"
    else
	round_opt=""
    fi
    # iff action != createAuction, add round test
    if [[ "${action}" != "createAuction" ]]; then
	round_text=" in round $round"
    else
	round_text=""
    fi

    # auction id handling
    if [[ "${action}" != "createAuction" ]] && [[ ! -z "$auctionId" ]]; then
	auction_opt="-auction-id ${auctionId}"
	auction_text=" of auction ${auctionId}"
    else
	auction_opt=""
	auction_text=""
    fi

    # execute request
    if [ ${do_it} -eq 1 ]; then
	say "user '$user' performs '$action'$round_text${auction_text}"
	result=$(${CLI} --user "${user}" --request "${action}" ${round_opt} ${auction_opt} ${scenario_opt1} "${scenario_opt2}") || die "failed to run request '${action}' as user '${user}'${round_text}${auction_text} (rc=$? /result='${result}')"

	# extract return code from status object (which is required to exist)
	rc=$(echo ${result} | sed "s/.*http-body='\(.*\)'/\1/g" | jq -r .status.rc) || die "returned result '${result}' does not contain a valid status object"
	[ ${rc} -eq ${expected_rc} ] || die "expected return code ${expected_rc} but got ${rc} instead in result '${result}"

	# try to extract auctionid if we created a new auction
	if [[ "${action}" == "createAuction" ]]; then
	    auctionId=$(echo $result | sed 's/.*auctionId\":\([0-9]*\)}.*/\1/g')
	fi
    else
	say "user '$user' would perform '$action'$round_text${auction_text}"
	${CLI} --dry-run --user "${user}" --request "${action}" ${round_opt} ${auction_opt} ${scenario_opt1} "${scenario_opt2}" || die "failed to run dry-run request '${action}' as user '${user}'${round_text}${auction_text} (rc=$? /result='${result}')"

	if [ ${wait} -eq 1 ]; then
	    wait "(Hit any key to continue)"
	fi
    fi
}


wait() {
    msg="$1"
    varName="$2"

    yell "${msg}"
    if [ ${non_interactive} -eq 0 ]; then
	eval read $varName
    fi
}

