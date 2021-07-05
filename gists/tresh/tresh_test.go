package tresh

import (
	"fmt"
	"github.com/ALTree/bigfloat"
	"github.com/spacemeshos/fixed"
	"math"
	"math/big"
	"math/rand"
	"testing"
)

func tFloat64(k,q,W float64) float64 {
	return 1.0 - math.Pow(2.0, -(float64(k)/(float64(W)*(1-float64(q)))))
}

func tFixed(k,q,W float64) fixed.Fixed {
	p := fixed.From(k).Div(fixed.One.Sub(fixed.From(q)).Mul(fixed.From(W))).Neg()
	// e^x = e^(y*ln(2)) = (e^ln(2))^y = 2^y
	return fixed.One.Sub(fixed.Exp(p.Mul(fixed.Ln2Value)))
}

func tBigfloat(k,q,W float64) float64 {
	x := big.NewFloat(-(float64(k)/(float64(W)*(1-float64(q)))))
	f, _ := bigfloat.Pow(big.NewFloat(2.0), x).Float64()
	return 1 - f
}

func Test1(t *testing.T) {

	const epsilon = 1e-12
	const epsilon2 = 1e-16 // use 1e-15 to pass test

	for k := 1; k < 100; k++ {
		for q := 2; q < 100; q++ {
			W := rand.Int()
			a := tFloat64(float64(k),float64(q)/3,float64(W))
			b := tFixed(float64(k),float64(q)/3,float64(W))
			c := tBigfloat(float64(k),float64(q)/3,float64(W))
			d := math.Abs(b.Float() - a)
			if d > epsilon {
				t.Fatalf("abs(%v - %v) = %v > %v", a,b.Float(), d, epsilon)
			}
			d2 := math.Abs(b.Float() - c)
			if d2 > epsilon2 {
				t.Errorf("abs(%v - %v) = %v > %v", c,b.Float(), d2, epsilon2)
			}
		}
	}

}

func Test2(t *testing.T) {
	k := 40.0
	q := 1.0/3
	W := 60.0
	a := tFloat64(k,q,W)
	fmt.Println(a)
	b := tFixed(k,q,W)
	fmt.Println(b.Float(),b)
	c := tBigfloat(k,q,W)
	fmt.Println(c)
}
