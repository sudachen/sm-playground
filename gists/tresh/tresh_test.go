package tresh

import (
	"github.com/spacemeshos/fixed"
	"math"
	"math/rand"
	"testing"
)

func tFloat64(k,q,W int) float64 {
	return 1.0 - math.Pow(2.0, -(float64(k)/(float64(W)*(1-float64(q)))))
}

func tFixed(k,q,W int) fixed.Fixed {
	p := fixed.New(k).Div(fixed.One.Sub(fixed.New(q)).Mul(fixed.New(W))).Neg()
	// e^x = e^(y*ln(2)) = (e^ln(2))^y = 2^y
	return fixed.One.Sub(fixed.Exp(p.Mul(fixed.Ln2Value)))
}

func Test1(t *testing.T) {

	const epsilon = 1e-12

	for k := 1; k < 100; k++ {
		for q := 2; q < 100; q++ {
			W := rand.Int()
			a := tFloat64(k,q,W)
			b := tFixed(k,q,W)
			d := math.Abs(b.Float() - a)
			if d > epsilon {
				t.Fatalf("abs(%v - %v) = %v > %v", a,b.Float(), d, epsilon)
			}
		}
	}

}
