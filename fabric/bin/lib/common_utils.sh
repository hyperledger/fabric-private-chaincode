# Copyright Intel Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

cred=`tput setaf 1`
cgrn=`tput setaf 2`
cblu=`tput setaf 4`
cmag=`tput setaf 5`
cwht=`tput setaf 7`
cbld=`tput bold`
bred=`tput setab 1`
bgrn=`tput setab 2`
bblu=`tput setab 4`
bwht=`tput setab 7`
crst=`tput sgr0`

function recho () {
    echo "${cbld}${cred}" $@ "${crst}" >&2
}

function becho () {
    echo "${cbld}${cblu}" $@ "${crst}" >&2
}

# Common reporting functions: say, yell & die
#-----------------------------------------
# say is stdout, yell is stderr
function say () {
    echo "$(basename $0): $*"
}


function yell () {
    becho "$(basename $0): $*" >&2;
}

function die() {
    recho "$(basename $0): $*" >&2
    exit 111
}

function para() {
    echo -e "\n"
}

# Common functions to run commands
#-----------------------------------------
try() {
    "$@" || die "test failed: $*"
}

# Fabric config parsing
#----------------------------
# input parameter is CONFIG_HOME directory
# when succesfull, will have defined following variables
# - SPID_FILE
# - API_KEY_FILE
# - NET_ID
# - PEER_ID
parse_fabric_config() {
    CONFIG_HOME=$1

    [ -d "${CONFIG_HOME}" ] || die "CONFIG_HOME not properly defined as '${CONFIG_HOME}'"
    [ -e "${CONFIG_HOME}/core.yaml" ] || die "no core.yaml in CONFIG_HOME '${CONFIG_HOME}'"
    SPID_FILE=$(perl -0777 -n -e 'm/spid:\s*file:\s*(\S+)/i && print "$1"' ${CONFIG_HOME}/core.yaml)
    API_KEY_FILE=$(perl -0777 -n -e 'm/apiKey:\s*file:\s*(\S+)/i && print "$1"' ${CONFIG_HOME}/core.yaml)
    PEER_ID=$(perl -0777 -n -e 'm/id:\s*(\S+)/i && print "$1"' ${CONFIG_HOME}/core.yaml)
    NET_ID=$(perl -0777 -n -e 'm/networkId:\s*(\S+)/i && print "$1"' ${CONFIG_HOME}/core.yaml)
}
