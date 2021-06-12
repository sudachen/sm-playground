package queries

import (
	"fmt"
	"testing"
)

var genesis = "202105161025"

func Test_AppStarted(t *testing.T) {
	r := GetAppStartedMsgs(genesis)
	fmt.Println(len(r))
}

func Test_Errors(t *testing.T) {
	r := FindErrors(genesis)
	fmt.Println(len(r))
}
