#!/bin/bash

LIBRARY_PATH=/opt/runnc/lib/
TARGET_BIN=bin/nabla_run

mkdir -p ${LIBRARY_PATH}

for i in $(ldd $(which whoami)); do 
	if [[ ${i} == /* ]]; then 
		echo Copying $i to ${LIBRARY_PATH}
        cp $i ${LIBRARY_PATH}/

        if [[ $? -ne 0 ]] ; then
            exit 1
        fi
	fi
done
