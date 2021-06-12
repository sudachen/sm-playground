package tutl

import (
	"fmt"
	"testing"
)

func Test_1(t *testing.T) {
	rs := query("new layer hash", "202105220732", 6)
	for _,r := range rs {
		fmt.Printf("%#v\n",r)
	}
}
