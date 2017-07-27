package gobconn

import (
	"reflect"
	"testing"
	"unsafe"
)

type T1 struct {
	Name string
	Age  int
}

type T2 struct {
	Job  bool
	Addr []string
	HOB  T1
	NiMa T3
}

type T3 struct {
	Name int
	Age  int
}

func Test_clearDate(t *testing.T) {
	value := T2{Job: true, Addr: []string{"1", "2"}, HOB: T1{"dijielin", 25}, NiMa: T3{12, 21}}
	test := reflect.ValueOf(&value)
	size := test.Elem().Type().Size()
	ClearData(size, unsafe.Pointer(test.Pointer()))
}

func Benchmark_clearDate(t *testing.B) {
	value := T2{Job: true, Addr: []string{"1", "2"}, HOB: T1{"dijielin", 25}, NiMa: T3{12, 21}}
	for i := 0; i < t.N; i++ {
		test := reflect.ValueOf(&value)
		ClearData(test.Elem().Type().Size(), unsafe.Pointer(test.Pointer()))
	}
}
