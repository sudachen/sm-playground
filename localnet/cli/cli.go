package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"os/signal"
	"sudachen.xyz/pkg/localnet"
	"sudachen.xyz/pkg/localnet/fu"
	"syscall"
)

const MajorVersion = 1
const MinorVersion = 0
const keyValueFormat = "%-8s %v\n"

var mainCmd = &cobra.Command{
	Use:           "localnet",
	Short:         fmt.Sprintf("Spacemesh LocalNet %v.%v (https://github.com/sudachen/sm-playground/localnet)", MajorVersion, MinorVersion),
	SilenceErrors: true,
}

var OptTrace = mainCmd.PersistentFlags().BoolP("trace", "x", false, "backtrace on panic")

func init() {
	mainCmd.PersistentFlags().BoolP("help", "h", false, "help for info")
	fu.VerboseOptP = mainCmd.PersistentFlags().BoolP("verbose", "v", false, "be verbose")
	mainCmd.AddCommand(
		cmdInfo,
		cmdStart,
		cmdStop,
	)
}

func localNet() *localnet.Localnet {
	return localnet.New()
}

func CLI() *cobra.Command {
	return mainCmd
}

func Main() {
	cst, _ := terminal.GetState(int(os.Stdin.Fd()))
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-c
		terminal.Restore(int(os.Stdin.Fd()), cst)
		os.Exit(1)
	}()
	if err := CLI().Execute(); err != nil {
		panic(err)
	}
}

