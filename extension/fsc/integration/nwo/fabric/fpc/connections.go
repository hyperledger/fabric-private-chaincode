/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/fabric/network"
	"github.com/pkg/errors"
)

const connectionsYaml = "connections.yaml"

const connectionsTemplate = `---
{{ with $w := . -}}
name: {{ .NetworkName }}
version: 1.0.0
client:
  organization: {{ .Name }}
  connection:
    timeout:
      peer:
        endorser: '300'

organizations:
  {{ .Name }}:
    mspid: {{ .MSPID }}
    peers: {{ range .PeersInOrg }}
      - {{ . }}
	{{- end }} 
    cryptoPath: {{ .MSPDir }}
{{- end }}

peers:{{ range .Peers }}
  {{- if .IsAnchor }}
  {{ .Name }}:
    url: grpcs://{{ .PeerAddr }}
    tlsCACerts:
      path: {{ .TLSDir }}/ca.crt
  {{- end }}
{{- end }}
`

type connections struct {
	NetworkName string
	Name        string
	MSPID       string
	PeersInOrg  []string
	MSPDir      string
	Peers       []peer
}

type peer struct {
	Name     string
	IsAnchor bool
	PeerAddr string
	TLSDir   string
}

// generateConnections collects the network information to produce a connection.yaml for a given organization `myOrg`.
// Returns an error if generation has failed. The connection.yaml is stored in the artifact folder `crypto/peerOrganization/orgDomain`.
func (n *Extension) generateConnections(myOrg string) error {
	logger.Infof("Generate connections for %s", myOrg)

	// get org
	org := n.network.Organization(myOrg)

	co := connections{
		NetworkName: fmt.Sprintf("network-%s-%s", n.network.NetworkID, org.Name),
		Name:        org.Name,
		MSPID:       org.MSPID,
		PeersInOrg:  []string{},
		MSPDir:      n.network.PeerOrgMSPDir(org),
		Peers:       []peer{},
	}

	// gather all peers of my org
	for _, peer := range n.network.PeersInOrg(myOrg) {
		co.PeersInOrg = append(co.PeersInOrg, fmt.Sprintf("%s.%s", peer.Name, org.Domain))
	}

	// gather anchor peers
	for _, p := range n.network.Peers {
		if !p.Anchor() {
			// we only want anchor peers
			continue
		}
		org := n.network.Organization(p.Organization)
		co.Peers = append(co.Peers, peer{
			Name:     fmt.Sprintf("%s.%s", p.Name, org.Domain),
			IsAnchor: p.Anchor(),
			PeerAddr: n.network.PeerAddress(p, network.ListenPort),
			TLSDir:   n.network.PeerLocalTLSDir(p),
		})
	}

	// set path
	p := filepath.Join(
		n.network.Context.RootDir(),
		n.network.Prefix,
		"crypto",
		"peerOrganizations",
		org.Domain,
		connectionsYaml,
	)

	if err := writeConnections(p, &co); err != nil {
		return errors.Wrapf(err, "creating connections profile failed")
	}

	return nil
}

// writeConnections applies the connectionsTemplate to a connections data structure and writes the result at a given path.
// Returns an error if parsing the template failed or the resulting connections file cannot be written to path.
func writeConnections(path string, co *connections) error {
	core, err := os.Create(path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file (%s)", path)
	}
	defer core.Close()

	t, err := template.New(connectionsYaml).Parse(connectionsTemplate)
	if err != nil {
		return errors.Wrapf(err, "failed to parse connections template")
	}

	err = t.Execute(io.Writer(core), co)
	if err != nil {
		return errors.Wrapf(err, "failed to apply data structure to template")
	}

	return nil
}
