all: test_docker_images node_tests.iso

node_tests.iso:
	genisoimage -l -r -o node_tests.iso testdata/hello

test_docker_images:
	sudo docker build -f Dockerfile.hello . -t nablact/nabla-hello:test
	sudo docker build -f Dockerfile.curl . -t nablact/nabla-curl:test
	sudo docker build -f Dockerfile.node . -t nablact/nabla-node:test
