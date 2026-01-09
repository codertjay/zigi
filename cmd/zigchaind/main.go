package main

import (
	"fmt"
	"os"

	clienthelpers "cosmossdk.io/client/v2/helpers"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"zigchain/app"
	"zigchain/cmd/zigchaind/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, clienthelpers.EnvPrefix, app.DefaultNodeHome); err != nil {
		_, err := fmt.Fprintln(rootCmd.OutOrStderr(), err)
		if err != nil {
			fmt.Printf("main failed to print error")
		}
		os.Exit(1)
	}
}
