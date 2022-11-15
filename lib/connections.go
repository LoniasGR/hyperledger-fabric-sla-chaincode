package lib

import (
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"github.com/hyperledger/fabric-gateway/pkg/identity"

)

// NewGrpcConnection creates a gRPC connection to the Gateway server.
func NewGrpcConnection(conf Config) (*grpc.ClientConn, error) {
	certificate, err := loadCertificate(conf.tlsCertPath)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "")

	connection, err := grpc.Dial(conf.peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	return connection, nil
}

// NewIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func NewIdentity(conf Config) *identity.X509Identity {
	log.Print(conf.UserConf.Credentials.Certificate)
	certificate, err := identity.CertificateFromPEM([]byte(conf.UserConf.Credentials.Certificate))
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(conf.UserConf.MspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

// NewSign creates a function that generates a digital signature from a message digest using a private key.
func NewSign(conf Config) identity.Sign {
	privateKeyPEM := conf.UserConf.Credentials.PrivateKey

	privateKey, err := identity.PrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}
