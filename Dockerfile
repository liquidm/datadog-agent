ARG GOVERSION=1.17.3
FROM registry.lqm.io/golang:${GOVERSION}
WORKDIR /var/app/datadog-agent

COPY . .

RUN apt-get update -y && \
    apt-get install -y curl jq netcat python3 python3-dev supervisor unzip wget

RUN wget https://bootstrap.pypa.io/get-pip.py && \
    python3 get-pip.py && \
    pip install -r requirements.txt

RUN invoke install-tools

CMD scripts/docker/cmd.sh
