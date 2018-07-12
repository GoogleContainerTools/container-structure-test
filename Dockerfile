FROM golang:1.10
ADD https://github.com/GoogleCloudPlatform/docker-credential-gcr/releases/download/v1.4.3-static/docker-credential-gcr_linux_amd64-1.4.3.tar.gz /usr/local/bin/
RUN tar -C /usr/local/bin/ -xvzf /usr/local/bin/docker-credential-gcr_linux_amd64-1.4.3.tar.gz

FROM gcr.io/distroless/base:latest
COPY --from=0 /usr/local/bin/docker-credential-gcr /docker-credential-gcr
ADD out/container-structure-test-linux-amd64 /container-structure-test
ENTRYPOINT ["/container-structure-test"]