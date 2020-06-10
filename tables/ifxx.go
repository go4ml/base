package tables

import (
	"fmt"
	"go-ml.dev/pkg/base/fu"
	"go-ml.dev/pkg/base/fu/lazy"
	"math"
	"reflect"
)

const epsilon = 1e-9

func equalf(vc reflect.Value) func(v reflect.Value) bool {
	switch vc.Kind() {
	case reflect.Slice:
		vv := []func(reflect.Value) bool{}
		for i := 0; i < vc.Len(); i++ {
			vv = append(vv, equalf(vc.Index(i)))
		}
		return func(v reflect.Value) bool {
			for _, f := range vv {
				if f(v) {
					return true
				}
			}
			return false
		}
	case reflect.Interface:
		return equalf(vc.Elem())
	case reflect.Float32, reflect.Float64:
		vv := vc.Float()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return math.Abs(v.Float()-vv) < epsilon
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return math.Abs(float64(v.Uint())-vv) < epsilon
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return math.Abs(float64(v.Int())-vv) < epsilon
			default:
				if v.Type() == fu.Fixed8Type {
					return math.Abs(float64(v.Interface().(fu.Fixed8).Float32())-vv) < epsilon
				}
			}
			return false
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vv := vc.Int()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return int64(v.Float()) == vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return int64(v.Uint()) == vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Int() == vv
			default:
				if v.Type() == fu.Fixed8Type {
					return int64(v.Interface().(fu.Fixed8).Float32()) == vv
				}
			}
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vv := vc.Uint()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return uint64(v.Float()) == vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return uint64(v.Uint()) == vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Uint() == vv
			default:
				if v.Type() == fu.Fixed8Type {
					return uint64(v.Interface().(fu.Fixed8).Float32()) == vv
				}
			}
			return false
		}
	case reflect.String:
		vv := vc.String()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return fmt.Sprintf("%d", v.Uint()) == vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return fmt.Sprintf("%d", v.Int()) == vv
			case reflect.String:
				return vv == v.String()
			}
			return false
		}
	default:
		return func(v reflect.Value) bool {
			return reflect.DeepEqual(v, vc)
		}
	}
}

func lessf(c interface{}) func(v reflect.Value) bool {
	vc := reflect.ValueOf(c)
	switch vc.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vv := vc.Int()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return int64(v.Float()) < vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return int64(v.Uint()) < vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Int() < vv
			}
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vv := vc.Uint()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return uint64(v.Float()) < vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return uint64(v.Uint()) < vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Uint() < vv
			}
			return false
		}
	case reflect.String:
		vv := vc.String()
		return func(v reflect.Value) bool {
			if v.Kind() == reflect.String {
				return vv < v.String()
			}
			return false
		}
	default:
		return func(v reflect.Value) bool {
			return fu.Less(v, vc)
		}
	}
}

func greatf(c interface{}) func(v reflect.Value) bool {
	vc := reflect.ValueOf(c)
	switch vc.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		vv := vc.Int()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return int64(v.Float()) > vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return int64(v.Uint()) > vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Int() > vv
			}
			return false
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		vv := vc.Uint()
		return func(v reflect.Value) bool {
			switch v.Kind() {
			case reflect.Float64, reflect.Float32:
				return uint64(v.Float()) > vv
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return uint64(v.Uint()) > vv
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return v.Uint() > vv
			}
			return false
		}
	case reflect.String:
		vv := vc.String()
		return func(v reflect.Value) bool {
			if v.Kind() > reflect.String {
				return vv < v.String()
			}
			return false
		}
	default:
		return func(v reflect.Value) bool {
			return fu.Less(vc, v)
		}
	}
}

func (zf Lazy) IfEq(c string, v ...interface{}) Lazy {
	var eq func(reflect.Value) bool
	if len(v) == 0 {
		return zf
	}
	if len(v) > 2 {
		eq = equalf(reflect.ValueOf(v))
	} else {
		eq = equalf(reflect.ValueOf(v[0]))
	}
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if eq(lr.ValueAt(j)) {
				return
			}
			return fu.True, nil
		}
	}
}

func (zf Lazy) TrueIfEq(c string, v interface{}, flag string) Lazy {
	eq := equalf(reflect.ValueOf(v))
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if eq(lr.ValueAt(j)) {
				lr.Set(flag, fu.True)
			} else {
				lr.Set(flag, fu.False)
			}
			return
		}
	}
}

func (zf Lazy) IfNe(c string, v interface{}) Lazy {
	eq := equalf(reflect.ValueOf(v))
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if !eq(lr.ValueAt(j)) {
				return
			}
			return fu.True, nil
		}
	}
}

func (zf Lazy) TrueIfNe(c string, v interface{}, flag string) Lazy {
	eq := equalf(reflect.ValueOf(v))
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if !eq(lr.ValueAt(j)) {
				lr.Set(flag, fu.True)
			} else {
				lr.Set(flag, fu.False)
			}
			return
		}
	}
}

func (zf Lazy) IfLt(c string, v interface{}) Lazy {
	lt := lessf(v)
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if lt(lr.ValueAt(j)) {
				return
			}
			return fu.True, nil
		}
	}
}

func (zf Lazy) IfGt(c string, v interface{}) Lazy {
	gt := greatf(v)
	return func() lazy.Stream {
		z := zf()
		nx := fu.AtomicSingleIndex{}
		return func(index uint64) (v reflect.Value, err error) {
			if v, err = z(index); err != nil || v.Kind() == reflect.Bool {
				return
			}
			lr := v.Interface().(fu.Struct)
			j, ok := nx.Get()
			if !ok {
				j, _ = nx.Set(lr.Pos(c))
			}
			if gt(lr.ValueAt(j)) {
				return
			}
			return fu.True, nil
		}
	}
}
