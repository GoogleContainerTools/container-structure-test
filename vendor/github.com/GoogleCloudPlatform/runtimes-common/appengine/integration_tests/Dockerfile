FROM gcr.io/google-appengine/python

ENV PATH=$PATH:/builder/google-cloud-sdk/bin/

RUN apt-get update && \
    apt-get install -y --force-yes wget unzip ca-certificates git && \
    # Setup Google Cloud SDK (latest)
    wget -nv https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-148.0.0-linux-x86_64.tar.gz && \
    mkdir -p /builder && \
    tar -xzf google-cloud-sdk-148.0.0-linux-x86_64.tar.gz -C /builder && \
    rm google-cloud-sdk-148.0.0-linux-x86_64.tar.gz && \
    /builder/google-cloud-sdk/install.sh --usage-reporting=false \
        --bash-completion=false \
        --disable-installation-options && \
    # Install alpha and beta components
    /builder/google-cloud-sdk/bin/gcloud -q components install alpha beta && \
    apt-get install -y --force-yes python-dev && \
    # Clean up
    apt-get remove -y --force-yes wget unzip && \
    apt-get clean


COPY requirements.txt /

RUN pip install --upgrade pip && pip install --upgrade -r /requirements.txt

COPY testsuite /testsuite

ENTRYPOINT ["/testsuite/driver.py"]
