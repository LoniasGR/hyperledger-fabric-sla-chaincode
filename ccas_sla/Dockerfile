FROM golang:1.18-alpine3.16 as builder

WORKDIR /go/src/github.com/LoniasGR/fabric-samples/hyperledger-fabric-sla-chaincode/ccas_sla

RUN apk add build-base

COPY . .

RUN go get -d -v ./...
RUN go build -tags musl -v ./...

EXPOSE 8999

FROM alpine:3.16

COPY --from=builder /go/src/github.com/LoniasGR/fabric-samples/hyperledger-fabric-sla-chaincode/ccas_sla/ccas_sla /usr/local/bin

EXPOSE 8999
CMD ["ccas_sla"]
