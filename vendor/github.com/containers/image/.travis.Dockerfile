FROM ubuntu:zesty

RUN apt-get -qq update && \
    apt-get install -y sudo docker.io git make golang btrfs-tools libdevmapper-dev libgpgme-dev libostree-dev
