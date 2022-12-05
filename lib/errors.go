package lib

import (
	"context"
	"errors"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"google.golang.org/grpc/status"
)

func handleError(err error) {
	if err != nil {
		switch err := err.(type) {
		case *client.EndorseError:
			log.Println(Red("Endorse error with gRPC status %v: %s\n", status.Code(err), err))
		case *client.SubmitError:
			log.Println(Red("Submit error with gRPC status %v: %s\n", status.Code(err), err))
		case *client.CommitStatusError:
			if errors.Is(err, context.DeadlineExceeded) {
				log.Println(Red("Timeout waiting for transaction %s commit status: %s", err.TransactionID, err))
			} else {
				log.Println(Red("Error obtaining commit status with gRPC status %v: %s\n", status.Code(err), err))
			}
		case *client.CommitError:
			log.Println(Red("Transaction %s failed to commit with status %d: %s\n", err.TransactionID, int32(err.Code), err))
		default:
			log.Println(Red("%v", err))
		}

		// Any error that originates from a peer or orderer node external to the gateway will have its details
		// embedded within the gRPC status error. The following code shows how to extract that.
		statusErr := status.Convert(err)
		for _, detail := range statusErr.Details() {
			switch detail := detail.(type) {
			case *gateway.ErrorDetail:
				log.Println(Red("Error from endpoint: %s, mspId: %s, message: %s\n", detail.Address, detail.MspId, detail.Message))
			}
		}
	}
}
