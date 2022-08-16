/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"

	"github.com/hyperledger/fabric-private-chaincode/samples/application/simple-cli-go/pkg"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(invokeCmd)
}

var invokeCmd = &cobra.Command{
	Use:   "invoke [function] [arg1] [arg2] ...",
	Short: "invoke FPC Chaincode with function and arguments",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := pkg.NewClient(config)
		res := client.Invoke(args[0], args[1:]...)
		fmt.Println("> " + res)
	},
}
