FROM golang:1.11
RUN go get -u github.com/golang/dep/cmd/dep
RUN apt update
RUN apt install -y genisoimage
RUN apt install -y libseccomp-dev
RUN apt install -y sudo
RUN apt install -y jq
