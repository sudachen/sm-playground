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

/*
Fnzi returns the first non zero value
*/
func Fnzi(a ...int) int {
	for _, i := range a {
		if i != 0 {
			return i
		}
	}
	return 0
}

/*
Fnzb returns the first non false value
*/
func Fnzb(a ...bool) bool {
	for _, i := range a {
		if i {
			return i
		}
	}
	return false
}

/*
MixMaps mixes count of map[string]string
*/
func MixMaps(a map[string]string, ms ...map[string]string) map[string]string {
	r := map[string]string{}
	for _, m := range append([]map[string]string{a},ms...) {
		for k,v := range m {
			r[k] = v
		}
	}
	return r
}

