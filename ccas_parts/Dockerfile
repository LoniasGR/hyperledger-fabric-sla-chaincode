FROM golang:1.18-alpine3.16 as builder

WORKDIR /go/src/github.com/LoniasGR/fabric-samples/hyperledger-fabric-sla-chaincode/ccas_parts

RUN apk add build-base

COPY . .

RUN go get -d -v ./...
RUN go build -tags musl -v ./...


FROM alpine:3.16

COPY --from=builder /go/src/github.com/LoniasGR/fabric-samples/hyperledger-fabric-sla-chaincode/ccas_parts/ccas_parts /usr/local/bin

EXPOSE 8999
CMD ["ccas_parts"]