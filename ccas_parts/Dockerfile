FROM golang:1.17-alpine3.16

WORKDIR /go/src/github.com/LoniasGR/fabric-samples/hyperledger-fabric-sla-chaincode/ccas_parts

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 9999
CMD ["ccas_parts"]
