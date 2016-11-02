FROM gcr.io/cloud-builders/docker

RUN mkdir /test
COPY structure_test /test/structure_test
VOLUME /test
COPY run.sh /run.sh

ENTRYPOINT ["/run.sh"]
