package tables

import (
	"fmt"
	"go-ml.dev/pkg/base/fu"
	"go-ml.dev/pkg/base/fu/lazy"
	"go-ml.dev/pkg/zorros"
	"math"
	"reflect"
	"sync"
)

type Lazy lazy.Source
type Sink lazy.Sink

func (Lazy) IsLazy() bool     { return true }
func (zf Lazy) Table() *Table { return zf.LuckyCollect() }
func (zf Lazy) Lazy() Lazy    { return zf }

func SourceError(err error) Lazy {
	return func() lazy.Stream {
		return func(_ uint64) (reflect.Value, error) {
			return reflect.Value{}, err
		}
	}
}

func SinkError(err error) Sink {
	return func(_ reflect.Value) error {
		return err
	}
}

func (zf Lazy) Map(f interface{}) Lazy {
	return func() lazy.Stream {
		z := zf()
		vf := reflect.ValueOf(f)
		vt := vf.Type()
		or, ir := vt, vt
		if vf.Kind() == reflect.Func {
			ir = vt.In(0)
			or = vt.Out(0)
		} else if vf.Kind() != reflect.Struct {
			panic("only func(struct{...})struct{...} and struct{...} is allowed as an argument of lazy.Map")
		}
		unwrap := fu.Unwrapper(ir)
		wrap := fu.Wrapper(or)
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return v, err
			}
			x := unwrap(v.Interface().(fu.Struct))
			if vf.Kind() == reflect.Func {
				x = vf.Call([]reflect.Value{x})[0]
			}
			return reflect.ValueOf(wrap(x)), nil
		}
	}
}

func (zf Lazy) Update(f interface{}) Lazy {
	return func() lazy.Stream {
		z := zf()
		vf := reflect.ValueOf(f)
		vt := vf.Type()
		or, ir := vt, vt
		if vf.Kind() == reflect.Func {
			ir = vt.In(0)
			or = vt.Out(0)
		} else if vf.Kind() != reflect.Struct {
			panic("only func(struct{...})struct{...} and struct{...} is allowed as an argument of lazy.Transform")
		}
		unwrap := fu.Unwrapper(ir)
		transform := fu.Transformer(or)
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return v, err
			}
			x := unwrap(v.Interface().(fu.Struct))
			if vf.Kind() == reflect.Func {
				x = vf.Call([]reflect.Value{x})[0]
			}
			return transform(x, v), nil
		}
	}
}

func (zf Lazy) Filter(f interface{}) Lazy {
	return func() lazy.Stream {
		z := zf()
		vf := reflect.ValueOf(f)
		vt := vf.Type()
		unwrap := fu.Unwrapper(vt.In(0))
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return v, err
			}
			x := unwrap(v.Interface().(fu.Struct))
			if vf.Call([]reflect.Value{x})[0].Bool() {
				return
			}
			return reflect.ValueOf(true), nil
		}
	}
}

func (zf Lazy) First(n int) Lazy {
	return Lazy(lazy.Source(zf).First(n))
}

func (zf Lazy) Parallel(concurrency ...int) Lazy {
	return Lazy(lazy.Source(zf).Parallel(concurrency...))
}

const iniCollectLength = 13
const maxChankLength = 10000

func (zf Lazy) Collect() (t *Table, err error) {
	length := 0
	columns := []reflect.Value{}
	names := []string{}
	na := []fu.Bits{}
	err = zf.Drain(func(v reflect.Value) error {
		if v.Kind() != reflect.Bool {
			lr := v.Interface().(fu.Struct)
			if length == 0 {
				names = lr.Names
				columns = make([]reflect.Value, len(names))
				na = make([]fu.Bits, len(names))
				for i, x := range lr.Columns {
					columns[i] = reflect.MakeSlice(reflect.SliceOf(x.Type()), 0, iniCollectLength)
				}
			}
			defer func() {
				if e := recover(); e != nil {
					fmt.Println(e)
					fmt.Println(lr.String())
					panic(e)
				}
			}()
			for i, x := range lr.Columns {
				if lr.Na.Bit(i) {
					columns[i] = reflect.Append(columns[i], reflect.Zero(columns[i].Type().Elem()))
					na[i].Set(length, true)
				} else {
					columns[i] = reflect.Append(columns[i], x)
				}
			}
			length++
		}
		return nil
	})
	if err != nil {
		return
	}
	return MakeTable(names, columns, na, length), nil
}

func (zf Lazy) LuckyCollect() *Table {
	t, err := zf.Collect()
	if err != nil {
		panic(zorros.Panic(err))
	}
	return t
}

func (zf Lazy) Drain(sink Sink) (err error) {
	return lazy.Source(zf).Drain(sink)
}

func (zf Lazy) LuckyDrain(sink Sink) {
	if err := zf.Drain(sink); err != nil {
		panic(zorros.Panic(err))
	}
}

func (zf Lazy) Count() (int, error) {
	return lazy.Source(zf).Count()
}

func (zf Lazy) LuckyCount() int {
	c, err := zf.Count()
	if err != nil {
		panic(zorros.Panic(err))
	}
	return c
}

func (zf Lazy) Rand(seed int, prob float64) Lazy {
	return Lazy(lazy.Source(zf).Rand(seed, prob))
}

func (zf Lazy) RandSkip(seed int, prob float64) Lazy {
	return Lazy(lazy.Source(zf).RandSkip(seed, prob))
}

func (zf Lazy) RandomFlag(c string, seed int, prob float64) Lazy {
	return func() lazy.Stream {
		z := zf()
		nr := fu.NaiveRandom{Value: uint32(seed)}
		wc := fu.WaitCounter{Value: 0}
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == lazy.STOP {
				wc.Stop()
			}
			if wc.Wait(index) {
				if err == nil && v.Kind() != reflect.Bool {
					lr := v.Interface().(fu.Struct)
					p := nr.Float()
					val := reflect.ValueOf(p < prob)
					v = reflect.ValueOf(lr.Set(c, val))
				}
				wc.Inc()
			}
			return
		}
	}
}

func (zf Lazy) Round(prec int) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lrx := v.Interface().(fu.Struct)
			lr := lrx.Copy(0)
			for i, c := range lr.Columns {
				switch c.Kind() {
				case reflect.Float32:
					lr.Columns[i] = reflect.ValueOf(fu.Round32(float32(c.Float()), prec))
				case reflect.Float64:
					lr.Columns[i] = reflect.ValueOf(fu.Round64(c.Float(), prec))
				}
			}
			return reflect.ValueOf(lr), nil
		}
	}
}

func (zf Lazy) IfFlag(c string) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			if j := fu.IndexOf(c, lr.Names); j >= 0 && lr.Columns[j].Bool() {
				return
			}
			return fu.True, nil
		}
	}
}

func (zf Lazy) IfNotFlag(c string) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			if j := fu.IndexOf(c, lr.Names); j < 0 || !lr.Columns[j].Bool() {
				return
			}
			return fu.True, nil
		}
	}
}

func (zf Lazy) Alias(c string, a string) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			return reflect.ValueOf(lr.Set(a, lr.Value(c))), nil
		}
	}
}

func (zf Lazy) True(c string) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			return reflect.ValueOf(lr.Set(c, fu.True)), nil
		}
	}
}

func (zf Lazy) False(c string) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			return reflect.ValueOf(lr.Set(c, fu.False)), nil
		}
	}
}

func (zf Lazy) Only(c ...string) Lazy {
	return func() lazy.Stream {
		z := zf()
		var only func(fu.Struct) fu.Struct
		mu := sync.Mutex{}
		f := fu.AtomicFlag{}
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			if !f.State() {
				mu.Lock()
				if !f.State() {
					only = fu.OnlyFilter(lr.Names, c...)
					f.Set()
				}
				mu.Unlock()
			}
			return reflect.ValueOf(only(lr)), nil
		}
	}
}

func (zf Lazy) Chain(zx Lazy) Lazy {
	return Lazy(lazy.Source(zf).Chain(lazy.Source(zx), func(a, b reflect.Value) (eqt bool) {
		if lr, ok := a.Interface().(fu.Struct); ok {
			if lrx, ok := b.Interface().(fu.Struct); ok {
				if len(lrx.Names) == len(lr.Names) {
					for i, n := range lrx.Names {
						if n != lr.Names[i] || lrx.Columns[i].Type() != lr.Columns[i].Type() {
							return false
						}
					}
					eqt = true
				}
			}
		}
		return
	}))
}

func (zf Lazy) Kfold(seed int, kfold int, k int, name string) Lazy {
	return func() lazy.Stream {
		z := zf()
		rnd := fu.NaiveRandom{Value: uint32(seed)}
		ac := fu.AtomicCounter{Value: 0}
		wc := fu.WaitCounter{Value: 0}
		nx := make([]int, kfold)
		for i := range nx {
			nx[i] = i
		}
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if index == lazy.STOP {
				wc.Stop()
			}
			if wc.Wait(index) {
				if err == nil && v.Kind() != reflect.Bool {
					a := int(ac.PostInc())
					if a%kfold == 0 {
						for i := range nx {
							j := int(rnd.Float() * float64(kfold))
							nx[i], nx[j] = nx[j], nx[i]
						}
					}
					lr := v.Interface().(fu.Struct)
					if nx[a%kfold] == k {
						v = reflect.ValueOf(lr.Set(name, fu.True))
					} else {
						v = reflect.ValueOf(lr.Set(name, fu.False))
					}
				}
				wc.Inc()
			}
			return
		}
	}
}

func (zf Lazy) Transform(f func(fu.Struct) (fu.Struct, bool, error)) Lazy {
	return func() lazy.Stream {
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			lr, ok, err := f(lr)
			if err != nil {
				return fu.False, err
			}
			if !ok {
				return fu.True, nil
			}
			return reflect.ValueOf(lr), nil
		}
	}
}

func (zf Lazy) BatchTransform(batch int, tf func(int) (FeaturesMapper, error)) Lazy {
	return zf.Batch(batch).Transform(tf).Flat()
}

func (zf Lazy) BatchReduce(batch int, tf func(*Table) (fu.Struct, bool, error)) Lazy {
	return zf.Batch(batch).Reduce(tf)
}

func (zf Lazy) Foreach(f func(fu.Struct) error) (err error) {
	return zf.Drain(func(v reflect.Value) error {
		if v.Kind() != reflect.Bool {
			lr := v.Interface().(fu.Struct)
			return f(lr)
		}
		return nil
	})
}

func (zf Lazy) UnpackTensor(c string) Lazy {
	return func() lazy.Stream {
		var ft func(fu.Struct) fu.Struct
		m := sync.Mutex{}
		fc := fu.AtomicFlag{Value: 0}
		z := zf()
		return func(index uint64) (v reflect.Value, err error) {
			v, err = z(index)
			if err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			if !fc.State() {
				m.Lock()
				if !fc.State() {
					ft = fu.TensorUnpacker(lr, c)
					fc.Set()
				}
				m.Unlock()
			}
			return reflect.ValueOf(ft(lr)), nil
		}
	}
}

func LazyConcatf(f ...func() Lazy) Lazy {
	return func() lazy.Stream {
		zf := f[0]()()
		fl := f[1:]
		wc := fu.WaitCounter{Value: 0}
		c := uint64(0)
		return func(index uint64) (v reflect.Value, err error) {
			if index == lazy.STOP {
				v, err = zf(index)
				wc.Stop()
				return
			}
			if wc.Wait(index) {
				v, err = zf(c)
				c++
				if err != nil {
					wc.Stop()
				} else {
					if v.Kind() == reflect.Bool && !v.Bool() {
						if len(fl) > 0 {
							_, _ = zf(lazy.STOP)
							zf = fl[0]()()
							fl = fl[1:]
							v = fu.True
							c = 0
						}
					}
					wc.Inc()
				}
			}
			return
		}
	}
}

func LazyConcat(a ...AnyData) Lazy {
	f := make([]func() Lazy, len(a))
	for i, x := range a {
		q := x
		f[i] = func() Lazy { return q.Lazy() }
	}
	return LazyConcatf(f...)
}

func LazyZip(a ...AnyData) Lazy {
	return func() lazy.Stream {
		zf := make([]lazy.Stream, len(a))
		vx := make([]fu.Struct, len(a))
		dx := make([]uint64, len(a))
		nx := []int{}
		names := []string{}
		for i, x := range a {
			zf[i] = x.Lazy()()
		}
		wc := fu.WaitCounter{Value: 0}
		return func(index uint64) (v reflect.Value, err error) {
			if index == lazy.STOP {
				for _, f := range zf {
					_, _ = f(index)
				}
				wc.Stop()
				return fu.False, nil
			}
			for wc.Wait(index) {
				for i, f := range zf {
				retry:
					for {
						v, err = f(dx[i])
						dx[i]++
						if err != nil || v.Kind() == reflect.Bool && !v.Bool() {
							wc.Stop()
							return fu.False, err
						}
						if v.Kind() != reflect.Bool {
							vx[i] = v.Interface().(fu.Struct)
							break retry
						}
					}
				}
				if len(names) == 0 {
					for _, a := range vx {
						for _, n := range a.Names {
							if fu.IndexOf(n, names) < 0 {
								nx = append(nx, len(names))
								names = append(names, n)
							} else {
								nx = append(nx, -1)
							}
						}
					}
				}
				k := 0
				columns := make([]reflect.Value, len(names))
				na := fu.Bits{}
				for _, a := range vx {
					for i, v := range a.Columns {
						j := nx[k]
						k++
						if j >= 0 {
							columns[j] = v
							na.Set(j, a.Na.Bit(i))
						}
					}
				}
				lr := fu.Struct{names, columns, na}
				wc.Inc()
				return reflect.ValueOf(lr), nil
			}
			return fu.False, nil
		}
	}
}

func (zf Lazy) Reals() ([]float32, error) {
	length := 0
	r := []float32{}
	err := zf.Drain(func(v reflect.Value) error {
		if v.Kind() != reflect.Bool {
			lr := v.Interface().(fu.Struct)
			if len(lr.Names) != 1 {
				return zorros.New("Lazy.Errors can handle only one column")
			}
			defer func() {
				if e := recover(); e != nil {
					fmt.Println(e)
					fmt.Println(lr.String())
					panic(e)
				}
			}()
			if lr.Na.Bit(0) {
				r = append(r, float32(math.NaN()))
			} else {
				r = append(r, fu.Cell{lr.Columns[0]}.Real())
			}
			length++
		}
		return nil
	})
	return r, err
}

func (zf Lazy) LuckyReals() []float32 {
	r, err := zf.Reals()
	if err != nil {
		panic(zorros.Panic(err))
	}
	return r
}
