package fu

import "math"

func Meanr(a []float32) float32 {
	var c float64
	for _, x := range a {
		c += float64(x)
	}
	return float32(c / float64(len(a)))
}

func Mse(a, b []float32) float32 {
	var c float64
	for i, x := range a {
		q := float64(x - b[i])
		c += q * q
	}
	return float32(c / float64(len(a)))
}

func Flatnr(a [][]float32) []float32 {
	n := 0
	for _, x := range a {
		n += len(x)
	}
	r := make([]float32, n)
	i := 0
	for _, x := range a {
		copy(r[i:i+len(x)], x)
		i += len(x)
	}
	return r
}

func MinMaxr(a []float32) (float32, float32) {
	min := a[0]
	max := a[0]
	for _, x := range a[1:] {
		if x > max {
			max = x
		} else if x < min {
			min = x
		}
	}
	return min, max
}

func Avgr(a []float32) float32 {
	var c float64
	for _, x := range a {
		c += float64(x)
	}
	return float32(c / float64(len(a)))
}

func Absr(a float32) float32 {
	if a >= 0 {
		return a
	}
	return -a
}

func Sigmar(a []float32) float64 {
	return math.Sqrt(Varr(a))
}

func Varr(a []float32) float64 {
	var m float64
	for _, x := range a {
		m += float64(x)
	}
	m /= float64(len(a))
	var s float64
	for _, x := range a {
		q := float64(x) - m
		s += q * q
	}
	s /= float64(len(a))
	return s
}
