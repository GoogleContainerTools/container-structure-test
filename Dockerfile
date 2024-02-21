FROM ubuntu:22.04
ADD out/container-structure-test /container-structure-test
ENTRYPOINT ["/container-structure-test"]
