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

function setup() {
	cd ${TOPLEVEL}/tests/integration
}

function nabla_run() {
	run ${TOPLEVEL}/build/runnc-cont \
		-nabla-run ${TOPLEVEL}/build/nabla-run "$@"

	echo "nabla-run $@ (status=$status):" >&2
	echo "$output" >&2
}

function docker_node_nabla_run() {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-node:test

	echo "nabla-run $@ (status=$status):" >&2
	echo "$output" >&2
}

@test "test hello" {
	nabla_run -unikernel test_hello.nabla
	[[ "$output" == *"Hello, World"* ]]
	[ "$status" -eq 0 ]
}

@test "test node-hello from a pre-built node_tests.iso" {
	nabla_run -unikernel node.nabla -volume node_tests.iso:/hello -- /hello/app.js
	[[ "$output" == *"hello from node"* ]]
	[ "$status" -eq 0 ]
}

@test "test node-hello runnc" {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-node:test /hello/app.js
	[[ "$output" == *"hello from node"* ]]
	[ "$status" -eq 0 ]
}

@test "test node-env" {
	nabla_run -unikernel node.nabla -env BLA=bla -env NABLA_ENV_TEST=blableblibloblu -env BLE=ble -volume node_tests.iso:/hello -- /hello/env.js
	echo "$output"
	[[ "$output" == *"env=blableblibloblu"* ]]
	[ "$status" -eq 0 ]
}

@test "test node-cwd" {
	nabla_run -unikernel node.nabla -cwd /hello -volume node_tests.iso:/hello -- /hello/cwd.js
	echo "$output"
	[[ "$output" == *"cwd=/hello"* ]]
	[ "$status" -eq 0 ]
}

@test "test hello with arg" {
	nabla_run -unikernel test_hello.nabla -- hola
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *"hola"* ]]
	[ "$status" -eq 0 ]
}

@test "test hello with json arg" {
	run nabla_run -unikernel test_hello.nabla -- {\"bla\":\"ble\"}
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *'{"bla":"ble"}'* ]]
	[ "$status" -eq 0 ]
}

@test "test hello with escaped json arg" {
	run nabla_run -unikernel test_hello.nabla -- '{\"bla\":\"ble\"}'
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *"{\\\"bla\\\":\\\"ble\\\"}"* ]]
	[ "$status" -eq 0 ]
}

@test "test hello runnc" {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-hello:test /test_hello.nabla
	[[ "$output" == *"Hello, World"* ]]
	[ "$status" -eq 0 ]
}

@test "test hello runnc with arg" {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-hello:test /test_hello.nabla hola
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *"hola"* ]]
	[ "$status" -eq 0 ]
}

@test "test hello runnc with json arg" {
	local-test

	run sudo docker run --rm --runtime=runnc nablact/nabla-hello:test /test_hello.nabla {\"bla\":\"ble\"}
	[[ "$output" == *"Hello, World"* ]]
	[[ "$output" == *"{\\\"bla\\\":\\\"ble\\\"}"* ]]
	[ "$status" -eq 0 ]
}

@test "test node-env runnc" {
	local-test

	# env.js just prints the NABLA_ENV_TEST environment variable
	run sudo docker run --rm --runtime=runnc -e NABLA_ENV_TEST=blableblibloblu nablact/nabla-node:test /hello/env.js
	echo "$output"
	[[ "$output" == *"env=blableblibloblu"* ]]
	[ "$status" -eq 0 ]
}

@test "test hello with net setting" {
	nabla_run -ipv4 10.0.0.2 -gwv4 10.0.0.1 -unikernel test_hello.nabla
	[[ "$output" == *"Hello, World"* ]]
	[ "$status" -eq 0 ]
}

@test "test curl local" {
	local-test

	(
		python test-http-server.py
	)&

	sleep 3

	HOSTIP=$( ip route get 1 | awk '{print $NF;exit}' )
	nabla_run -ipv4 10.0.0.2 -gwv4 10.0.0.1 \
			-unikernel test_curl.nabla -- "$HOSTIP"

	[[ "$output" == *"XXXXXXXXXX"* ]]
	[ "$status" -eq 0 ]
}

@test "test curl runnc" {
	local-test

	(
		python test-http-server.py
	)&

	sleep 3

	HOSTIP=$( ip route get 1 | awk '{print $NF;exit}' )
	run sudo docker run --rm --runtime=runnc nablact/nabla-curl:test /test_curl.nabla "$HOSTIP"

	echo "$output"
	[[ "$output" == *"XXXXXXXXXX"* ]]
	[ "$status" -eq 0 ]
}

@test "test memory runnc" {
	local-test

	# Check that 1024m is passed correct to runnc as 1024.
	run sudo docker run --rm --runtime=runnc -m 1024m nablact/nabla-node:test /hello/app.js
	cat /proc/$!/cmdline
	[[ "$output" == *"--mem=1024"* ]]
	[ "$status" -eq 0 ]
}
