FROM ubuntu:16.04 as clone


FROM lumjjb/user-node:v0
COPY --from=clone /bin/sh /bin/sh
