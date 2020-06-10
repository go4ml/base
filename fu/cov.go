package fu

import "math"

func Corr(x, y []float32) float64 {
	var X, X2, Y2, Y, XY float64
	for i := range x {
		u := float64(x[i])
		v := float64(y[i])
		X += u
		Y += v
		X2 += u * u
		Y2 += v * v
		XY += u * v
	}
	n := float64(len(x))
	cov := (n*XY - X*Y)
	if math.Abs(cov) < 1e-12 {
		return 0
	}
	return cov / math.Sqrt((n*X2-X*X)*(n*Y2-Y*Y))
}

func Cord(x, y []float64) float64 {
	var X, X2, Y2, Y, XY float64
	for i := range x {
		u := x[i]
		v := y[i]
		X += u
		Y += v
		X2 += u * u
		Y2 += v * v
		XY += u * v
	}
	n := float64(len(x))
	return (n*XY - X*Y) / math.Sqrt((n*X2-X*X)*(n*Y2-Y*Y))
}
