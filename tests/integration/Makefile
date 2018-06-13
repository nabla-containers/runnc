all:
	sudo docker build -f Dockerfile.hello . -t nablact/nabla-hello:test
	sudo docker build -f Dockerfile.curl . -t nablact/nabla-curl:test
	sudo docker build -f Dockerfile.node . -t nablact/nabla-node:test
	genisoimage -l -r -o node_tests.iso hello
