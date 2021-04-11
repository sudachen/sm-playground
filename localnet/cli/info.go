package cli

import (
	"github.com/spf13/cobra"
)

var cmdInfo = &cobra.Command{
	Use:   "info",
	Short: "Display the localnet status",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
	},
}

