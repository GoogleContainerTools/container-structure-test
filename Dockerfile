FROM gcr.io/cloud-builders/docker

COPY structure_test /test/structure_test
COPY run.sh /run.sh

ENTRYPOINT ["/run.sh"]
