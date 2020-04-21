package fu

func Mean(a []float32) float32 {
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
