/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/sample/cli_app/pkg"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(queryCmd)
}

var queryCmd = &cobra.Command{
	Use:   "query [function] [arg1] [arg2] ...",
	Short: "query FPC Chaincode with function and arguments",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := pkg.NewClient(config)
		res := client.Call(args[0], args[1:]...)
		fmt.Println("> " + res)
	},
}
