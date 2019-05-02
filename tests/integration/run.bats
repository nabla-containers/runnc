#!/usr/bin/env bats
## Copyright (c) 2018, IBM
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

load helpers

RUNNC_OUT="out"

function setup() {
	setup_root
}

function teardown() {
	teardown_root
}

function docker_node_nabla_run() {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-node:test /node.nabla

	echo "nabla-run $* (status=$status):" >&2
	echo "$output" >&2
}

function runnc_run() {
	daemon=${2:-nodaemon}
	# trick: don't start with 'run', or you will get a deadlock
	# waiting for the nabla program to be done but no one start it
	runnc create --bundle "$TEST_BUNDLE" --pid-file "$ROOT/pid" "$1" > $RUNNC_OUT
	run runnc start "$1"

	echo "nabla-run $* (status=$status):" >&2
	echo "$output" >&2

	# waiting for container process to be done
	if [[ ${daemon} == 'nodaemon' ]]; then
		tail --pid="$(cat "$ROOT"/pid)" -f /dev/null
	fi
}

@test "hello" {
	setup_test "hello"
	local name="test-nabla-hello"

	config_mod '.process.args |= .+ ["test_hello.nabla"]'

	runnc_run "$name"

	run cat "$TEST_BUNDLE/$RUNNC_OUT"
	[[ "$output" == *"Hello, World"* ]]

	runnc delete --force "$name"
	teardown_test
}

@test "hello with arg" {
	setup_test "hello"
	local name="test-nabla-hello-arg"

	config_mod '.process.args |= .+ ["test_hello.nabla", "hola"]'

	runnc_run "$name"

	run cat "$TEST_BUNDLE/$RUNNC_OUT"
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *"hola"* ]]

	runnc delete --force "$name"
	teardown_test
}

@test "hello with json arg" {
	setup_test "hello"
	local name="test-nabla-hello-json-arg"

	config_mod '.process.args |= .+ ["test_hello.nabla"]'
	config_mod '.process.args |= .+ ["{\"bla\":\"ble\"}"]'

	runnc_run "$name"

	run cat "$TEST_BUNDLE/$RUNNC_OUT"
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *'{\"bla\":\"ble\"}'* ]]

	runnc delete --force "$name"
	teardown_test
}

@test "hello with escaped json arg" {
	setup_test "hello"
	local name="test-nabla-hello-escaped-json-arg"

	config_mod '.process.args |= .+ ["test_hello.nabla"]'
	config_mod '.process.args |= .+ ["{\\\"bla\\\":\\\"ble\\\"}"]'

	runnc_run "$name"

	run cat "$TEST_BUNDLE/$RUNNC_OUT"
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *'{\\\"bla\\\":\\\"ble\\\"}'* ]]

	runnc delete --force "$name"
	teardown_test
}

@test "hello with net setting" {
	skip "TODO: Require proper networking for native runnc in prestart hooks"
}

@test "hello runnc" {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-hello:test /test_hello.nabla
	[[ "$output" == *"Hello, World"* ]]
}

@test "hello runnc with arg" {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-hello:test /test_hello.nabla hola
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *"hola"* ]]
}

@test "hello runnc with json arg" {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-hello:test /test_hello.nabla "{\"bla\":\"ble\"}"
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *"{\\\"bla\\\":\\\"ble\\\"}"* ]]
}

@test "node hello" {
	setup_test "node"
	local name="test-nabla-node"

	config_mod '.process.args |= .+ ["node.nabla", "/hello/app.js"]'

	runnc_run "$name"

	run cat "$TEST_BUNDLE/$RUNNC_OUT"
	[[ "$output" == *"hello from node"* ]]

	runnc delete --force "$name"
	teardown_test
}

@test "node env" {
	setup_test "node"
	local name="test-nabla-node-env"

	config_mod '.process.args |= .+ ["node.nabla", "/hello/env.js"]'
	config_mod '.process.env |= .+ ["BLA=bla", "NABLA_ENV_TEST=blableblibloblu", "BLE=ble"]'

	runnc_run "$name"

	run cat "$TEST_BUNDLE/$RUNNC_OUT"
	[[ "$output" == *"env=blableblibloblu"* ]]

	runnc delete --force "$name"
	teardown_test
}

@test "node cwd" {
	setup_test "node"
	local name="test-nabla-node-cwd"

	config_mod '.process.args |= .+ ["node.nabla", "/hello/cwd.js"]'
	config_mod '.process.cwd |= "/hello"'

	runnc_run "$name"

	run cat "$TEST_BUNDLE/$RUNNC_OUT"
	[[ "$output" == *"cwd=/hello"* ]]

	runnc delete --force "$name"
	teardown_test
}

@test "node hello runnc" {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-node:test /node.nabla /hello/app.js
	[[ "$output" == *"hello from node"* ]]
}

@test "node env runnc" {
	local-test

	# env.js just prints the NABLA_ENV_TEST environment variable
	run sudo docker run --rm --runtime=runnc -e NABLA_ENV_TEST=blableblibloblu nablact/nabla-node:test /node.nabla /hello/env.js
	[[ "$output" == *"env=blableblibloblu"* ]]
}

@test "curl local" {
	skip "TODO: Require proper networking for native runnc in prestart hooks"
}

@test "curl runnc" {
	local-test

	(
		python "$TESTDATA"/test-http-server.py
	)&

	sleep 3

	HOSTIP=$( ip route get 1 | awk '{print $NF;exit}' )
	run sudo docker run --rm --runtime=runnc nablact/nabla-curl:test /test_curl.nabla "$HOSTIP"

	echo "$output"
	[[ "$output" == *"XXXXXXXXXX"* ]]
	[ "$status" -eq 0 ]
}

@test "memory runnc" {
	local-test

	# Check that 1024m is passed correct to runnc as 1024.
	# Redirecting stderr to dev/null because there is a kernel warning
	memory_check() {
	  sudo docker run -d --rm --runtime=runnc -m 1024m nablact/nabla-node:test /node.nabla /hello/app.js 2>/dev/null
	}

	run memory_check
	[ "$status" -eq 0 ]
	container_pid=$(docker inspect --format '{{.State.Pid}}' "${output}")
	run bash -c "sudo ps -e -o pid,command | grep ${container_pid}"
	[ "$status" -eq 0 ]
	[[ "$output" == *"--mem=1024"* ]]
}

@test "oom_adjust" {
	setup_test "node"
	local name="test-nabla-oom-adjust"

	config_mod '.process.args |= .+ ["node.nabla", "/hello/app.js"]'
	config_mod '.process.oomScoreAdj |= .+ 104'

	runnc_run "${name}" "daemon"

	run bash -c "cat /proc/$(cat "${ROOT}"/pid)/oom_score_adj"
	[[ "$output" == *"104"* ]]

	runnc delete --force "${name}"
	teardown_test
}
