#!/bin/bash
#
# Copyright Intel Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

# assumes FPC_ROOT and DEMO_CLIENT_SCRIPTS_DIR is defined

. "${FPC_ROOT}"/fabric/bin/lib/common_utils.sh


CLI="${DEMO_CLIENT_SCRIPTS_DIR}/cli"




#  Simple "DSL" to script auction scenarios
#---------------------------------------------
#
# TODO (maybe)
# - get auction status after submit and output when rounds/auction is finished?
#
# terms of language are
# - scenario [<scenario-path>] if path is missing, take directory where script file is located
# - "submit" <user> <action>
# - "wait" <message> [<variable-for-user-input>]


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
    say "user '$user' performs '$action'$round_text${auction_text}"
    result=$(${CLI} -user "${user}" -request "${action}" ${round_opt} ${auction_opt} ${scenario_opt1} "${scenario_opt2}") || die "failed to run request '${action}' as user '${user}'${round_text}${auction_text} (rc=$? /result='${result}')"

    # try to extract auctionid if we created a new auction
    if [[ "${action}" == "createAuction" ]]; then
	auctionId=$(echo $result | sed 's/.*auctionId\":\([0-9]*\)}.*/\1/g')
    fi
}


wait() {
    msg="$1"
    varName="$2"

    yell "${msg}"
    eval read $varName
}

