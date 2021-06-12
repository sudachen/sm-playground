package fu

import "sudachen.xyz/pkg/localnet/stdio"

var VerboseOpt = false
var VerboseOptP *bool = &VerboseOpt

func Verbose(f string, a ...interface{}) {
	if VerboseOptP != nil && *VerboseOptP {
		stdio.Printfln("# "+f, a...)
	}
}

func Error(f string, a ...interface{}) {
	stdio.Errorf("! "+f+"\n", a...)
}
