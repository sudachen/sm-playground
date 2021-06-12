package cli

import (
	"github.com/spf13/cobra"
)


var instancesNumber int
var layersPerEpoch int
var layerDuration int
var reportPerf bool

func StartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start localnet",
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			l := localNet()
			if instancesNumber != 0 {
				l.Count = instancesNumber
			}
			if layerDuration != 0 {
				l.LayerDuration = layerDuration
			}
			if layersPerEpoch != 0 {
				l.LayersPerEpoch = layersPerEpoch
			}
			l.ReportPerf = reportPerf
			if err := l.Start(); err != nil {
				panic(err)
			}
		},
	}
	cmd.Flags().IntVarP(&instancesNumber, "instances", "n", 0, "Count of instances to start")
	cmd.Flags().IntVarP(&layersPerEpoch, "layers-per-epoch", "l", 0, "Layers per epoch")
	cmd.Flags().IntVarP(&layerDuration, "layers-duration", "d", 0, "Layer duration in seconds")
	cmd.Flags().BoolVarP(&reportPerf, "report-pscope", "r", false, "Report to PyroScope")
	return cmd
}
