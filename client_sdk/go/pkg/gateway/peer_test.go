/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-private-chaincode/internal/protoutil"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// endorser records the proposal it receives and answers with a fixed response
type endorser struct {
	peer.UnimplementedEndorserServer
	received *peer.SignedProposal
	response *peer.Response
}

func (e *endorser) ProcessProposal(_ context.Context, sp *peer.SignedProposal) (*peer.ProposalResponse, error) {
	e.received = sp
	return &peer.ProposalResponse{Response: e.response}, nil
}

func TestEvaluateAtPeer(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "user"}, NotAfter: time.Now().Add(time.Hour)}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	id, err := identity.NewX509Identity("Org1MSP", cert)
	if err != nil {
		t.Fatal(err)
	}
	sign, err := identity.NewPrivateKeySign(key)
	if err != nil {
		t.Fatal(err)
	}

	// the enclave peer, reachable only through dial
	e := &endorser{response: &peer.Response{Status: 200, Payload: []byte("result")}}
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	peer.RegisterEndorserServer(srv, e)
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	var dialed string
	dial := func(endpoint string) (grpc.ClientConnInterface, error) {
		dialed = endpoint
		return grpc.NewClient("passthrough:///bufnet",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// the gateway connection is never used, the proposal is only built by the gateway contract
	gwConn, err := grpc.NewClient("passthrough:///unused", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	gw, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(gwConn))
	if err != nil {
		t.Fatal(err)
	}
	defer gw.Close()
	contract := gw.GetNetwork("mychannel").GetContract("mycc")

	result, err := EvaluateAtPeer(contract, sign, dial, "peer1.org1:7051", "__invoke", []byte("arg"))
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "result" || dialed != "peer1.org1:7051" {
		t.Fatalf("got result %q from %q", result, dialed)
	}

	// the peer receives the invocation, signed by the client identity
	prop, err := protoutil.UnmarshalProposal(e.received.GetProposalBytes())
	if err != nil {
		t.Fatal(err)
	}
	payload, err := protoutil.UnmarshalChaincodeProposalPayload(prop.GetPayload())
	if err != nil {
		t.Fatal(err)
	}
	spec, err := protoutil.UnmarshalChaincodeInvocationSpec(payload.GetInput())
	if err != nil {
		t.Fatal(err)
	}
	if args := spec.GetChaincodeSpec().GetInput().GetArgs(); len(args) != 2 || string(args[0]) != "__invoke" || string(args[1]) != "arg" {
		t.Fatalf("unexpected args %q", args)
	}
	digest := sha256.Sum256(e.received.GetProposalBytes())
	if !ecdsa.VerifyASN1(&key.PublicKey, digest[:], e.received.GetSignature()) {
		t.Fatal("proposal signature does not verify")
	}

	// an error response of the peer is returned as error
	e.response = &peer.Response{Status: 500, Message: "no enclave"}
	if _, err := EvaluateAtPeer(contract, sign, dial, "peer1.org1:7051", "__invoke"); err == nil {
		t.Fatal("expected error for status 500")
	}
}
