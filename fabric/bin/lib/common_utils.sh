# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

if [[ -z "$TERM" ]] || [[ "$TERM" == 'dumb' ]]; then
    # to avoid 'tput: No value for $TERM and no -T specified' errors ..
    cred=
    cgrn=
    cblu=
    cmag=
    cwht=
    cbld=
    bred=
    bgrn=
    bblu=
    bwht=
    crst=
else
    cred=$(tput setaf 1)
    cgrn=$(tput setaf 2)
    cblu=$(tput setaf 4)
    cmag=$(tput setaf 5)
    cwht=$(tput setaf 7)
    cbld=$(tput bold)
    bred=$(tput setab 1)
    bgrn=$(tput setab 2)
    bblu=$(tput setab 4)
    bwht=$(tput setab 7)
    crst=$(tput sgr0)
fi

function recho () {
    echo "${cbld}${cred}"$@"${crst}" >&2
}

function becho () {
    echo "${cbld}${cblu}"$@"${crst}" >&2
}

function gecho () {
    echo "${cbld}${cgrn}"$@"${crst}" >&2
}

# Common reporting functions: say, yell & die
#-----------------------------------------
# they all write to stderr. if you want normal progres for stdout, just use echo
function say () {
    echo "$(basename $0): $*" >&2;
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

try_fail() {
    recho "failure expected next"
    (! "$@") || die "rev-test failed: $*"
}

# Variant of try which stores commands stdout and stderr (or only stdout) in variable RESPONSE
try_r() {
    say "$@"
    export RESPONSE=$("$@" 2>&1) RESPONSE_TYPE="out+err" || die "test failed: $*"
}

try_out_r() {
    say "$@"
    export RESPONSE=$("$@") RESPONSE_TYPE="out" || die "test failed: $*"
}

