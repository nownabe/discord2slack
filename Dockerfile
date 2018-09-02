FROM golang:1.9-stretch


ENV DEBIAN_FRONTEND noninteractive
ENV LANG C.UTF-8


RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    python-pip \
    openjdk-8-jre-headless \
  && pip install grpcio-tools grpcio \
  && rm -rf /var/lib/apt/lists/*


# Install gcloud

ENV gcloud_version 214.0.0
ENV PATH /opt/google-cloud-sdk/bin:$PATH

RUN mkdir -p /opt \
  && curl -LSfs -o /opt/gcloud.tar.gz \
    https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${gcloud_version}-linux-x86_64.tar.gz \
  && tar xf /opt/gcloud.tar.gz -C /opt \
  && rm /opt/gcloud.tar.gz \
  && /opt/google-cloud-sdk/install.sh \
    --usage-reporting false \
    --rc-path $HOME/.bashrc \
    --command-completion true \
    --path-update true \
    --quiet \
  && gcloud components install app-engine-go --quiet \
  && gcloud components install cloud-datastore-emulator --quiet
