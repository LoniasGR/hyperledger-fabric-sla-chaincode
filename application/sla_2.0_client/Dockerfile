# syntax=docker/dockerfile:1

FROM golang:1.18

RUN apt-get update && apt-get install -y jq && rm -rf /var/lib/apt/lists/*

WORKDIR /go/src/github.com/LoniasGR/fabric-samples/hyperledger-fabric-sla-chaincode/sla_2.0_client

RUN curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/bootstrap.sh \
    | bash -s -- -s -d

# remove sample config files extracted by the installation script
RUN rm config/configtx.yaml && \
    rm config/core.yaml && \
    rm config/orderer.yaml

COPY . .
RUN chmod +x scripts/cc.sh

ENV PATH=bin:$PATH

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8999
CMD ["sla_2.0_client", "-f", "consumer.properties"]
