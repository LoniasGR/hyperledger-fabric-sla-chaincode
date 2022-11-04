FROM golang:1.18

WORKDIR /go/src/github.com/LoniasGR/fabric-samples/hyperledger-fabric-sla-chaincode/application

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8999
CMD ["application", "-f", "consumer.properties"]
