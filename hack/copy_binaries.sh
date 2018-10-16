#!/bin/bash

# Copyright (c) 2018, IBM
# Author(s): Brandon Lum, Ricardo Koller, Dan Williams
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

BIN_PATH=/usr/local/bin/

# We add binaries like runnc-cont and nabla-run to /opt/X since they are not 
# to be consumed directly by the user.
BIN_PATH2=/opt/runnc/bin/

COPY_BINS=("build/runnc" "build/runnc-cont" "build/nabla-run")

mkdir -p ${BIN_PATH2}

for i in ${COPY_BINS[@]}; do
    echo $i
    cp $i ${BIN_PATH}/
    cp $i ${BIN_PATH2}/
done
