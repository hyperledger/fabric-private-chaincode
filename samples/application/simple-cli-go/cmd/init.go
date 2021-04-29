/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"github.com/hyperledger/fabric-private-chaincode/samples/application/simple-cli-go/pkg"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init [target_peer]",
	Short: "initialize enclave",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		admin := pkg.NewAdmin(config)
		defer admin.Close()
		return admin.InitEnclave(args[0])
	},
}
