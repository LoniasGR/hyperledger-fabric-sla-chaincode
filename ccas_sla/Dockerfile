FROM golang:1.18-alpine3.16

WORKDIR /go/src/github.com/LoniasGR/fabric-samples/hyperledger-fabric-sla-chaincode/ccas_sla

RUN apk add build-base

COPY . .

RUN go get -d -v ./...
RUN go install -tags musl -v ./...

EXPOSE 8999
CMD ["ccas_sla"]
