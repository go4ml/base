package fu

import (
	"math"
	"sort"
)

type rankarr struct {
	d []float32
	n []int
}

func (a rankarr) Len() int {
	return len(a.d)
}
func (a rankarr) Swap(i, j int) {
	a.d[i], a.d[j] = a.d[j], a.d[i]
	a.n[i], a.n[j] = a.n[j], a.n[i]
}

func (a rankarr) Less(i, j int) bool {
	return a.d[i] < a.d[j]
}

func Rank(a []float32, first ...bool) []float32 {
	r := make([]float32, len(a))
	n := make([]int, len(a))
	for i, v := range a {
		r[i] = Round32(v, 4)
		n[i] = i
	}
	sort.Stable(rankarr{r, n})
	for i := 0; i < len(r); {
		j := i
		for i++; i < len(r) && r[j] == r[i]; i++ {
		}
		x := float32(j)
		if !Fnzb(first...) {
			x = float32(i-j)/2 + x
		}
		for ; j < i+1 && j < len(r); j++ {
			r[j] = x
		}
	}
	rank := make([]float32, len(a))
	for i, j := range n {
		rank[j] = r[i]
	}
	return rank
}

func RankPct(a []float32, first ...bool) []float32 {
	rank := Rank(a, first...)
	max := Maxr(float32(math.Inf(-1)), rank...)
	for i, x := range rank {
		rank[i] = x / max
	}
	return rank
}
