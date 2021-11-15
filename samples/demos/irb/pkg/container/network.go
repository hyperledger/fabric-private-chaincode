/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package container

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type Network struct {
	Name string
	id   string
}

func (n *Network) Create() error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	opts := types.NetworkCreate{
		Driver: "bridge",
	}
	resp, err := cli.NetworkCreate(ctx, n.Name, opts)
	if err != nil {
		return err
	}

	n.id = resp.ID

	fmt.Printf("Network created: %s (%s)\n", n.Name, resp.ID)
	return nil
}

func (n *Network) Remove() error {
	if len(n.id) == 0 {
		// no network to remove
		return nil
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	return cli.NetworkRemove(ctx, n.id)
}
