package fu

import (
	"math"
	"os"
	"sync/atomic"
	"time"
)

type NaiveRandom struct {
	Value uint32
}

func (nr *NaiveRandom) Reseed() {
	atomic.StoreUint32(&nr.Value, uint32(time.Now().UnixNano()+int64(os.Getpid())))
}

func (nr *NaiveRandom) Uint32() (r uint32) {
	for {
		r = atomic.LoadUint32(&nr.Value)
		rx := r
		r = r*1664525 + 1013904223
		if atomic.CompareAndSwapUint32(&nr.Value, rx, r) {
			break
		}
	}
	return
}

func (nr *NaiveRandom) Uint() uint {
	return uint(nr.Uint32())
}

func (nr *NaiveRandom) Int() int {
	return int(nr.Uint32())
}
func (nr *NaiveRandom) Float() float64 {
	return float64(nr.Uint()) / (float64(math.MaxUint32) + 1)
}

func Seed(seed int) int {
	if seed != 0 {
		return seed
	}
	return int(time.Now().UnixNano() + int64(os.Getpid()))
}

func Seed32(seed int) uint32 {
	return uint32(Seed(seed))
}

func Seed64(seed int) int64 {
	return int64(Seed(seed))
}

func RandomInts(seed int, count int) []int {
	nr := NaiveRandom{Value: uint32(seed)}
	m := map[int]bool{}
	for i := 0; i < count; {
		v := nr.Int()
		if _, ok := m[v]; !ok {
			m[v] = true
			i++
		}
	}
	r := make([]int, count)
	i := 0
	for k := range m {
		r[i] = k
		i++
	}
	return r
}
