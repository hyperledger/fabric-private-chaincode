/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package container

import (
	"bufio"
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Container struct {
	Image       string
	CMD         []string
	Name        string
	HostIP      string
	HostPort    string
	containerID string
	Env         []string
	Network     string
}

func (c *Container) Start() error {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: c.Image,
			Cmd:   c.CMD,
			Env:   c.Env,
		},
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port(fmt.Sprintf("%s/tcp", c.HostPort)): []nat.PortBinding{{HostIP: c.HostIP, HostPort: c.HostPort}},
			},
		}, nil, nil, c.Name)
	if err != nil {
		return err
	}

	if err := cli.NetworkConnect(context.Background(), c.Network, resp.ID, nil); err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	go func() {
		reader, err := cli.ContainerLogs(context.Background(), resp.ID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Timestamps: false,
		})
		defer reader.Close()
		if err != nil {
			return
		}

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			fmt.Printf("%s: %s", c.Name, scanner.Text())
		}
	}()
	c.containerID = resp.ID
	fmt.Printf("Container started: %s (%s)\n", c.Name, resp.ID)

	return nil
}

func (c *Container) Stop() error {
	if len(c.containerID) == 0 {
		// no container started
		return nil
	}

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	if err := cli.ContainerStop(ctx, c.containerID, nil); err != nil {
		return err
	}

	if err := cli.ContainerRemove(ctx, c.containerID, types.ContainerRemoveOptions{}); err != nil {
		return err
	}

	fmt.Printf("Container stopped: %s\n", c.containerID)

	return nil
}
