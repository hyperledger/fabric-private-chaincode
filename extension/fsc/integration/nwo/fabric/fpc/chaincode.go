/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"encoding/json"
	"fmt"

	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/core/contract"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/commands"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/network"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/topology"
)

type ChannelProvider struct {
	Network *network.Network
	CC      *topology.ChannelChaincode
}

func (c *ChannelProvider) GetContract(id string) fpc.Contract {
	// search the chaincode
	switch id {
	case c.CC.Chaincode.Name:
		return &ChannelContract{
			Network:     c.Network,
			Channel:     c.CC.Channel,
			ChaincodeID: id,
			CC:          c.CC,
		}
	case "ercc":
		for _, chaincode := range c.Network.Chaincodes(c.CC.Channel) {
			if chaincode.Chaincode.Name == "ercc" {
				return &ChannelContract{
					Network:     c.Network,
					Channel:     chaincode.Channel,
					ChaincodeID: id,
					CC:          chaincode,
				}
			}
		}
		panic(fmt.Sprintf("cannot find chaincode id [%s]", id))
	default:
		panic(fmt.Sprintf("unspected chaincode id [%s]", id))
	}
}

type ChannelTransaction struct {
	Network      *network.Network
	ChaincodeID  string
	Channel      string
	CC           *topology.ChannelChaincode
	FunctionName string
}

func (c *ChannelTransaction) Evaluate(args ...string) ([]byte, error) {
	org := c.Network.PeerOrgs()[0]
	peer := c.Network.Peer(org.Name, c.Network.PeersInOrgWithOptions(org.Name, false)[0].Name)
	s := &struct {
		Args []string `json:"Args,omitempty"`
	}{}
	s.Args = append(s.Args, c.FunctionName)
	for _, arg := range args {
		s.Args = append(s.Args, string(arg))
	}
	ctor, err := json.Marshal(s)
	Expect(err).NotTo(HaveOccurred())

	sess, err := c.Network.PeerUserSession(peer, "User1", commands.ChaincodeQuery{
		ChannelID: c.Channel,
		Name:      c.CC.Chaincode.Name,
		Ctor:      string(ctor),
	})
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, c.Network.EventuallyTimeout).Should(gexec.Exit(0))

	return sess.Buffer().Contents(), nil
}

type ChannelContract struct {
	Network     *network.Network
	ChaincodeID string
	Channel     string
	CC          *topology.ChannelChaincode
}

func (c *ChannelContract) Name() string {
	return c.CC.Chaincode.Name
}

func (c *ChannelContract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	org := c.Network.PeerOrgs()[0]
	peer := c.Network.Peer(org.Name, c.Network.PeersInOrgWithOptions(org.Name, false)[0].Name)
	s := &struct {
		Args []string `json:"Args,omitempty"`
	}{}
	s.Args = append(s.Args, name)
	for _, arg := range args {
		s.Args = append(s.Args, string(arg))
	}
	ctor, err := json.Marshal(s)
	Expect(err).NotTo(HaveOccurred())

	sess, err := c.Network.PeerUserSession(peer, "User1", commands.ChaincodeQuery{
		ChannelID: c.Channel,
		Name:      c.CC.Chaincode.Name,
		Ctor:      string(ctor),
	})
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, c.Network.EventuallyTimeout).Should(gexec.Exit(0))

	return sess.Buffer().Contents(), nil
}

func (c *ChannelContract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	orderer := c.Network.Orderer("orderer")
	org := c.Network.PeerOrgs()[0]
	peer := c.Network.Peer(org.Name, c.Network.PeersInOrgWithOptions(org.Name, false)[0].Name)
	s := &struct {
		Args []string `json:"Args,omitempty"`
	}{}
	s.Args = append(s.Args, name)
	for _, arg := range args {
		s.Args = append(s.Args, string(arg))
	}
	ctor, err := json.Marshal(s)
	Expect(err).NotTo(HaveOccurred())

	// Peers
	initOrgs := map[string]bool{}
	var peerAddresses []string
	peers := c.Network.PeersByName(c.CC.Peers)
	for _, p := range peers {
		if exists := initOrgs[p.Organization]; !exists {
			peerAddresses = append(peerAddresses, c.Network.PeerAddress(p, network.ListenPort))
			initOrgs[p.Organization] = true
		}
	}

	sess, err := c.Network.PeerUserSession(peer, "User1", commands.ChaincodeInvoke{
		ChannelID:     c.Channel,
		Orderer:       c.Network.OrdererAddress(orderer, network.ListenPort),
		Name:          c.CC.Chaincode.Name,
		Ctor:          string(ctor),
		PeerAddresses: peerAddresses,
	})
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, c.Network.EventuallyTimeout).Should(gexec.Exit(0))
	Expect(sess.Err).To(gbytes.Say("Chaincode invoke successful. result: status:200"))

	return sess.Buffer().Contents(), nil
}

func (c *ChannelContract) CreateTransaction(name string, peerEndpoints ...string) (fpc.Transaction, error) {
	return &ChannelTransaction{
		Network:      c.Network,
		ChaincodeID:  c.ChaincodeID,
		Channel:      c.Channel,
		CC:           c.CC,
		FunctionName: name,
	}, nil
}

type ChannelClient struct {
	n         *Extension
	peer      *topology.Peer
	orderer   *topology.Orderer
	chaincode *topology.ChannelChaincode
}

func (c *ChannelClient) Query(chaincodeID string, fcn string, args [][]byte, targetEndpoints ...string) ([]byte, error) {
	ci := &ChaincodeInput{
		Args: append(append([]string{}, fcn), ArgsToStrings(args)...),
	}
	logger.Infof("query [%s] with args...", chaincodeID)
	for i, arg := range ci.Args {
		logger.Infof("arg [%d][%s]", i, arg)
	}
	ctor, err := json.Marshal(ci)
	Expect(err).ToNot(HaveOccurred())

	sess, err := c.n.network.PeerUserSession(c.peer, "User1", commands.ChaincodeQuery{
		ChannelID: c.chaincode.Channel,
		Name:      chaincodeID,
		Ctor:      string(ctor),
	})
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, c.n.network.EventuallyTimeout).Should(gexec.Exit(0))

	return sess.Buffer().Contents(), nil
}

func (c *ChannelClient) Execute(chaincodeID string, fcn string, args [][]byte) (string, error) {
	ci := &ChaincodeInput{
		Args: append(append([]string{}, fcn), ArgsToStrings(args)...),
	}
	ctor, err := json.Marshal(ci)
	Expect(err).ToNot(HaveOccurred())

	initOrgs := map[string]bool{}
	var erccPeerAddresses []string
	peers := c.n.network.PeersByName(c.n.FPCERCC.Peers)
	for _, p := range peers {
		if exists := initOrgs[p.Organization]; !exists {
			erccPeerAddresses = append(erccPeerAddresses, c.n.network.PeerAddress(p, network.ListenPort))
			initOrgs[p.Organization] = true
		}
	}

	sess, err := c.n.network.PeerUserSession(c.peer, "User1", commands.ChaincodeInvoke{
		NetworkPrefix: c.n.network.Prefix,
		ChannelID:     c.chaincode.Channel,
		Orderer:       c.n.network.OrdererAddress(c.orderer, network.ListenPort),
		Name:          chaincodeID,
		Ctor:          string(ctor),
		PeerAddresses: erccPeerAddresses,
		WaitForEvent:  true,
		ClientAuth:    c.n.network.ClientAuthRequired,
	})
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess, c.n.network.EventuallyTimeout).Should(gexec.Exit(0))
	for i := 0; i < len(erccPeerAddresses); i++ {
		Eventually(sess.Err, c.n.network.EventuallyTimeout).Should(gbytes.Say(`\Qcommitted with status (VALID)\E`))
	}
	Expect(sess.Err).To(gbytes.Say("Chaincode invoke successful. result: status:200"))

	return "", nil
}

func ArgsToStrings(args [][]byte) []string {
	var res []string
	for _, arg := range args {
		res = append(res, string(arg))
	}
	return res
}
