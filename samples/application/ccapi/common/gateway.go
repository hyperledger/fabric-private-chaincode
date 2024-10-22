package common

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var (
	gatewayTLSCredentials *credentials.TransportCredentials
)

func CreateGrpcConnection(endpoint string) (*grpc.ClientConn, error) {
	// Check TLS credential was created
	if gatewayTLSCredentials == nil {
		gatewayServerName := os.Getenv("FABRIC_GATEWAY_NAME")

		cred, err := createTransportCredential(GetTLSCACert(), gatewayServerName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create tls credentials")
		}

		gatewayTLSCredentials = &cred
	}

	// Create client grpc connection
	return grpc.Dial(endpoint, grpc.WithTransportCredentials(*gatewayTLSCredentials))
}

func CreateGatewayConnection(grpcConn *grpc.ClientConn, user string) (*client.Gateway, error) {
	// Create identity
	id, err := newIdentity(getSignCert(user), GetMSPID())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new identity")
	}
	gatewayId := id

	// Create sign function
	sign, err := newSign(getSignKey(user))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new sign function")
	}

	gatewaySign := sign

	// Create a Gateway connection for a specific client identity.
	return client.Connect(
		gatewayId,
		client.WithSign(gatewaySign),
		client.WithClientConnection(grpcConn),

		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
}

// Create transport credential
func createTransportCredential(tlsCertPath, serverName string) (credentials.TransportCredentials, error) {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	return credentials.NewClientTLSFromCert(certPool, serverName), nil
}

// Creates a client identity for a gateway connection using an X.509 certificate.
func newIdentity(certPath, mspID string) (*identity.X509Identity, error) {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		return nil, err
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		return nil, err
	}

	return id, nil
}

// Creates a function that generates a digital signature from a message digest using a private key.
func newSign(keyPath string) (identity.Sign, error) {
	privateKeyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read private key file")
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create signer function")
	}

	return sign, nil
}

// Returns error and status code
func ParseError(err error) (error, int) {
	var errMsg string

	switch err := err.(type) {
	case *client.EndorseError:
		errMsg = "endorse error for transaction"
	case *client.SubmitError:
		errMsg = "submit error for transaction"
	case *client.CommitStatusError:
		if errors.Is(err, context.DeadlineExceeded) {
			errMsg = "timeout waiting for transaction commit status"
		} else {
			errMsg = "error obtaining commit status for transaction"
		}
	case *client.CommitError:
		errMsg = "transaction failed to commit"
	default:
		errMsg = "unexpected error type:" + err.Error()
	}

	statusErr := status.Convert(err)

	details := statusErr.Details()
	if len(details) == 0 {
		return errors.New(errMsg), http.StatusInternalServerError
	}

	for _, detail := range details {
		switch detail := detail.(type) {
		case *gateway.ErrorDetail:
			status, msg := extractStatusAndMessage(detail.Message)
			return errors.New(msg), status
		}
	}

	return errors.New(errMsg), http.StatusInternalServerError
}

func extractStatusAndMessage(msg string) (int, string) {
	pattern := `chaincode response (\b(\d{3})\b), `
	reg := regexp.MustCompile(pattern)
	matches := reg.FindStringSubmatch(msg)

	if len(matches) == 0 {
		return http.StatusInternalServerError, msg
	}

	errMsg := strings.Replace(msg, matches[0], "", 1)
	status, err := strconv.Atoi(matches[1])
	if err != nil {
		status = http.StatusInternalServerError
	}

	return status, errMsg
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}
	return identity.CertificateFromPEM(certificatePEM)
}

func getSignCert(user string) string {
	cryptoPath := GetCryptoPath()
	filename := user + "@" + os.Getenv("ORG") + "." + os.Getenv("DOMAIN") + "-cert.pem"

	return strings.Replace(cryptoPath, "{username}", user, 1) + "/signcerts/" + filename
}

func getSignKey(user string) string {
	cryptoPath := GetCryptoPath()
	filename := "priv_sk"

	return strings.Replace(cryptoPath, "{username}", user, 1) + "/keystore/" + filename
}
