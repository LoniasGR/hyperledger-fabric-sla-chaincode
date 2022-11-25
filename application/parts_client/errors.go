package main

import (
	"context"
	"errors"
	"log"

	"github.com/LoniasGR/hyperledger-fabric-sla-chaincode/lib"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"google.golang.org/grpc/status"
)

func handleError(err error) {
	if err != nil {
		switch err := err.(type) {
		case *client.EndorseError:
			log.Println(lib.Red("Endorse error with gRPC status %v: %s\n", status.Code(err), err))
		case *client.SubmitError:
			log.Println(lib.Red("Submit error with gRPC status %v: %s\n", status.Code(err), err))
		case *client.CommitStatusError:
			if errors.Is(err, context.DeadlineExceeded) {
				log.Println(lib.Red("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err))
			} else {
				log.Println(lib.Red("Error obtaining commit status with gRPC status %v: %s\n", status.Code(err), err))
			}
		case *client.CommitError:
			log.Println(lib.Red("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err))
		default:
			log.Println(lib.Red("%v", err))
		}

		// Any error that originates from a peer or orderer node external to the gateway will have its details
		// embedded within the gRPC status error. The following code shows how to extract that.
		statusErr := status.Convert(err)
		for _, detail := range statusErr.Details() {
			switch detail := detail.(type) {
			case *gateway.ErrorDetail:
				log.Println(lib.Red("Error from endpoint: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message))
			}
		}
	}
}
