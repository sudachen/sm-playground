package cli

import (
	"github.com/spf13/cobra"
)

var cmdStart = &cobra.Command{
	Use:   "start",
	Short: "Start localnet",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := localNet().Start(); err != nil {
			panic(err)
		}
	},
}

