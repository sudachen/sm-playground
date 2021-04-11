package cli

import "github.com/spf13/cobra"

var cmdStop = &cobra.Command{
	Use:   "stop",
	Short: "Stop localnet",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := localNet().Stop(); err != nil {
			panic(err)
		}
	},
}


