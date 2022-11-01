FROM golang:1.18

WORKDIR /go/src/github.com/LoniasGR/fabric-samples/hyperledger-fabric-sla-chaincode/application/vru_client

# RUN apk add --no-cache gcc musl-dev

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 9999
CMD ["vru_client", "-f", "consumer.properties"]
