#!/bin/bash
# Copyright (c) 2018, IBM
# Author(s): Brandon Lum, Ricardo Koller
#
# Permission to use, copy, modify, and/or distribute this software for
# any purpose with or without fee is hereby granted, provided that the
# above copyright notice and this permission notice appear in all
# copies.
#
# THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL
# WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED
# WARRANTIES OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE
# AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL
# DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA
# OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
# TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
# PERFORMANCE OF THIS SOFTWARE.

TOPLEVEL=$( dirname "${BASH_SOURCE[0]}" )

RUNNC="$TOPLEVEL"/../../build/runnc

# Test data path.
TESTDATA="$TOPLEVEL/testdata"

TEST_BUNDLE="$BATS_TMPDIR/test-nabla"

# Root state path.
ROOT="$BATS_TMPDIR/runnc.root"

# Wrapper for runnc.
function runnc() {
    "$RUNNC" --root "$ROOT" "$@"
}

function setup_test() {
    local test_type=$1

    run mkdir "$TEST_BUNDLE"
    cp "$TESTDATA"/config.json "$TEST_BUNDLE"

    case "$test_type" in
        "hello")
            cp "$TOPLEVEL"/test_hello.nabla "$TEST_BUNDLE"
            ;;
        "node")
            cp -r "$TESTDATA"/hello "$TEST_BUNDLE"/hello
            cp "$TOPLEVEL"/node.nabla "$TEST_BUNDLE"
            ;;
        "curl")
            cp "$TOPLEVEL"/test_curl.nabla "$TEST_BUNDLE"
            ;;
        *)
            echo "error: unknown test type: \"$test_type\""
            exit 1
            ;;
    esac
    cd "$TEST_BUNDLE" || exit
}

function config_mod () {
    # shellcheck disable=SC2002
    cat config.json | jq "$@" > config.json.new
    mv config.json.new config.json
}

function teardown_test() {
    run rm  -r "$TEST_BUNDLE"
}

function setup_root() {
    run mkdir -p "$ROOT"
}
function teardown_root() {
    run rm -r "$ROOT"
}

function local-test () {
     if [ -n "${INCONTAINER}" ]; then
         skip "Test cannot be run in container"
     fi
}

function container-test () {
     if [ -z "${INCONTAINER}" ]; then
         skip "Test must be run in container"
     fi
}
