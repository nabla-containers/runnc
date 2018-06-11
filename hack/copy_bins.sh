#!/bin/bash

BIN_PATH=/usr/local/bin/
COPY_BINS=("bin/runnc" "bin/runnc-cont" "bin/ukvm-bin")

for i in ${COPY_BINS[@]}; do
    echo $i
    cp $i ${BIN_PATH}/
done
