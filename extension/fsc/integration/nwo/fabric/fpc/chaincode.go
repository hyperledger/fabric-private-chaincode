/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"encoding/json"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/commands"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/network"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/topology"
)

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
	logger.Debugf("query [%s] with args %v", chaincodeID, ci.Args)
	ctor, err := json.Marshal(ci)
	if err != nil {
		return nil, err
	}

	sess, err := c.n.network.PeerUserSession(c.peer, "User1", commands.ChaincodeQuery{
		ChannelID: c.chaincode.Channel,
		Name:      chaincodeID,
		Ctor:      string(ctor),
	})
	if err != nil {
		return nil, err
	}
	Eventually(sess, c.n.network.EventuallyTimeout).Should(gexec.Exit(0))

	return sess.Buffer().Contents(), nil
}

func (c *ChannelClient) Execute(chaincodeID string, fcn string, args [][]byte) (string, error) {
	ci := &ChaincodeInput{
		Args: append(append([]string{}, fcn), ArgsToStrings(args)...),
	}
	ctor, err := json.Marshal(ci)
	if err != nil {
		return "", err
	}

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
	if err != nil {
		return "", err
	}
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
