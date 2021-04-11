package main

import (
	"fmt"
	"os"
	"sudachen.xyz/pkg/errstr"
	"sudachen.xyz/pkg/localnet/cli"
)

func main() {

	defer func() {
		if !*cli.OptTrace {
			if e := recover(); e != nil {
				fmt.Fprintln(os.Stderr, errstr.MessageOf(e))
				os.Exit(1)
			}
		}
	}()

	cli.Main()
}
