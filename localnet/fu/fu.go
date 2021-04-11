package fu


/*
Maxi returns maximal int value
*/
func Maxi(a int, b ...int) int {
	q := a
	for _, x := range b {
		if x > q {
			q = x
		}
	}
	return q
}

/*
Mini returns minimal int value
*/
func Mini(a int, b ...int) int {
	q := a
	for _, x := range b {
		if x < q {
			q = x
		}
	}
	return q
}
