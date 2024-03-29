package fu

import (
	"fmt"
	"go4ml.xyz/zorros"
	"reflect"
	"strings"
	"sync"
)

type Struct struct {
	Names   []string
	Columns []reflect.Value
	Na      Bits
}

func (lr Struct) String() string {
	r := make([]string, len(lr.Names))
	for i, n := range lr.Names {
		v := (interface{})("<nil>")
		if lr.Columns[i].IsValid() {
			v = lr.Columns[i].Interface()
		}
		r[i] = fmt.Sprintf("%v:%v", n, Ife(lr.Na.Bit(i), "N/A", v))
	}
	return "fu.Struct{" + strings.Join(r, ", ") + "}"
}

func (lr Struct) Copy(extra int) Struct {
	width := len(lr.Names)
	columns := make([]reflect.Value, width, width+extra)
	copy(columns, lr.Columns)
	names := make([]string, width, width+extra)
	copy(names, lr.Names)
	na := lr.Na.Copy()
	return Struct{names, columns, na}
}

func (lrx Struct) With(lr Struct) (r Struct) {
	extra := 0
	ndx := make([]int, len(lr.Names))
	for i, n := range lr.Names {
		j := IndexOf(n, lrx.Names)
		ndx[i] = j
		if j < 0 {
			extra++
		}
	}
	r = lrx.Copy(extra)
	for i, j := range ndx {
		if j >= 0 {
			r.Columns[j] = lr.Columns[i]
			r.Na.Set(j, lr.Na.Bit(i))
		} else {
			r.Na.Set(len(r.Names), lr.Na.Bit(i))
			r.Names = append(r.Names, lr.Names[i])
			r.Columns = append(r.Columns, lr.Columns[i])
		}
	}
	return
}

func Wrapper(rt reflect.Type) func(reflect.Value) Struct {
	L := rt.NumField()
	names := make([]string, L)
	for i := range names {
		names[i] = rt.Field(i).Name
	}
	return func(v reflect.Value) Struct {
		lr := Struct{Columns: make([]reflect.Value, L), Names: names, Na: Bits{}}
		for i := range names {
			x := v.Field(i)
			lr.Na.Set(i, Isna(x))
			lr.Columns[i] = x
		}
		return lr
	}
}

var uwrpMu = sync.Mutex{}

func Unwrapper(v reflect.Type) func(lr Struct) reflect.Value {
	var indecies [][]int
	inif := AtomicFlag{0}
	return func(lr Struct) reflect.Value {
		if !inif.State() {
			uwrpMu.Lock()
			if !inif.State() {
				var nd [][]int
				L := v.NumField()
				for i := 0; i < L; i++ {
					vt := v.Field(i)
					pat := string(vt.Tag)
					if pat == "" {
						pat = vt.Name
					}
					like := Pattern(pat)
					q := []int{}
					for i, n := range lr.Names {
						if like(n) {
							q = append(q, i)
						}
					}
					if len(q) == 0 {
						uwrpMu.Unlock()
						panic(zorros.Panic(zorros.Errorf("Struct does not have filed(s) matched to " + pat)))
					}
					if vt.Type.Kind() == reflect.Slice {
						nd = append(nd, q)
					} else {
						nd = append(nd, q[:1])
					}
				}
				indecies = nd
				inif.Set()
			}
			uwrpMu.Unlock()
		}

		x := reflect.New(v).Elem()
		for i, nd := range indecies {
			vt := v.Field(i)
			if vt.Type.Kind() == reflect.Slice {
				et := vt.Type.Elem()
				a := reflect.MakeSlice(reflect.SliceOf(et), len(nd), len(nd))
				for j, k := range nd {
					a.Index(j).Set(Convert(lr.Columns[k], lr.Na.Bit(k), et))
				}
				x.Field(i).Set(a)
			} else {
				k := nd[0]
				y := Convert(lr.Columns[k], lr.Na.Bit(k), vt.Type)
				x.Field(i).Set(y)
			}
		}
		return x
	}
}

var trfMu = sync.Mutex{}

func Transformer(rt reflect.Type) func(reflect.Value, reflect.Value) reflect.Value {
	var (
		names  []string
		update []int
	)
	inif := AtomicFlag{0}
	return func(v reflect.Value, olr reflect.Value) reflect.Value {
		lrx := olr.Interface().(Struct)
		if !inif.State() {
			trfMu.Lock()
			if !inif.State() {
				names = make([]string, len(lrx.Names), len(lrx.Names)*2)
				update = make([]int, len(lrx.Names), len(lrx.Names)*2)
				copy(names, lrx.Names)
				for i := range update {
					update[i] = -1
				}
				L := rt.NumField()
				for i := 0; i < L; i++ {
					n := rt.Field(i).Name
					if j := IndexOf(n, names); j < 0 {
						names = append(names, n)
						update = append(update, i)
					} else {
						update[j] = i
					}
				}
				inif.Set()
			}
			trfMu.Unlock()
		}
		lr := Struct{Columns: make([]reflect.Value, len(names)), Names: names, Na: lrx.Na.Copy()}
		for i := range names {
			if j := update[i]; j >= 0 {
				x := v.Field(j)
				lr.Na.Set(i, Isna(x))
				lr.Columns[i] = x
			} else {
				lr.Columns[i] = lrx.Columns[i]
			}
		}
		return reflect.ValueOf(lr)
	}
}

func NaStruct(names []string, tp reflect.Type) Struct {
	columns := make([]reflect.Value, len(names))
	for i := range columns {
		columns[i] = reflect.Zero(tp)
	}
	return Struct{names, columns, FillBits(len(names))}
}

func MakeStruct(names []string, vals ...interface{}) Struct {
	columns := make([]reflect.Value, len(names))
	for i := range columns {
		columns[i] = reflect.ValueOf(vals[i])
	}
	return Struct{names, columns, Bits{}}
}

func (lr Struct) Set(c string, val reflect.Value) Struct {
	cj := IndexOf(c, lr.Names)
	lr = lr.Copy(cj + 1)
	if cj < 0 {
		lr.Names = append(lr.Names, c)
		lr.Columns = append(lr.Columns, val)
	} else {
		lr.Columns[cj] = val
		lr.Na.Set(cj, false)
	}
	return lr
}

func (lr Struct) Pos(c string) int {
	return IndexOf(c, lr.Names)
}

func (lr Struct) ValueAt(i int) reflect.Value {
	return lr.Columns[i]
}

func (lr Struct) Value(c string) reflect.Value {
	j := IndexOf(c, lr.Names)
	return lr.Columns[j]
}

func (lr Struct) Index(c string) Cell {
	j := IndexOf(c, lr.Names)
	return Cell{lr.Columns[j]}
}

func (lr Struct) Int(c string) int       { return lr.Index(c).Int() }
func (lr Struct) Float(c string) float64 { return lr.Index(c).Float() }
func (lr Struct) Real(c string) float32  { return lr.Index(c).Real() }
func (lr Struct) Text(c string) string   { return lr.Index(c).Text() }

func (lr Struct) Round(p int) Struct {
	c := lr.Copy(0)
	for i, v := range c.Columns {
		switch v.Kind() {
		case reflect.Float32, reflect.Float64:
			c.Columns[i] = reflect.ValueOf(Round64(v.Float(), p))
		}
	}
	return c
}

func OnlyFilter(names []string, c ...string) func(Struct) Struct {
	ns := make([]string, 0, len(names))
	nx := make([]int, 0, len(names))
	p := make([]func(string) bool, len(c))
	for i, s := range c {
		p[i] = Pattern(s)
	}
	for i, n := range names {
	l:
		for _, f := range p {
			if f(n) {
				ns = append(ns, n)
				nx = append(nx, i)
				break l
			}
		}
	}
	return func(lr Struct) Struct {
		columns := make([]reflect.Value, len(ns))
		for i, x := range nx {
			columns[i] = lr.Columns[x]
		}
		return Struct{Names: ns, Columns: columns}
	}
}

func tensorUnpacker(names []string, c string, volume int) func(lr Struct) Struct {
	j := IndexOf(c, names)
	ns := make([]string, len(names)-1+volume)
	k := 0
	for i, n := range names {
		if i != j {
			ns[k] = n
			k++
		}
	}
	for i := 1; k < len(ns); k++ {
		ns[k] = fmt.Sprintf("%v%v", c, i)
		i++
	}
	k = len(names) - 1
	return func(lr Struct) Struct {
		t := lr.ValueAt(j).Interface().(Tensor)
		columns := make([]reflect.Value, len(lr.Names)-1+t.Volume())
		na := Bits{}
		t.Extract(columns[k:])
		n := 0
		for i, v := range lr.Columns {
			if i != j {
				if lr.Na.Bit(i) {
					na.Set(j, true)
				}
				columns[n] = v
				n++
			}
		}
		return Struct{ns, columns, na}
	}
}

func TensorUnpacker(lr Struct, c string) func(lr Struct) Struct {
	return tensorUnpacker(lr.Names, c, lr.Value(c).Interface().(Tensor).Volume())
}

func (lr Struct) UnpackTensor(c string) Struct {
	return TensorUnpacker(lr, c)(lr)
}
