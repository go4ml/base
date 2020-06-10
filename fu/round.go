package fu

func Round64(v float64, n int) float64 {
	k := float64(10)
	for i := 1; i < n; i++ {
		k *= 10
	}
	return float64(int64(v*k)) / k
}

func Round64s(v []float64, n int) []float64 {
	r := make([]float64, len(v))
	for i, x := range v {
		r[i] = Round64(x, n)
	}
	return r
}

func Round32(v float32, n int) float32 {
	k := float32(10)
	for i := 1; i < n; i++ {
		k *= 10
	}
	return float32(int32(v*k)) / k
}

func Round32s(v []float32, n int) []float32 {
	r := make([]float32, len(v))
	for i, x := range v {
		r[i] = Round32(x, n)
	}
	return r
}
