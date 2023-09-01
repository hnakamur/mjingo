package internal

import (
	"log"
	"reflect"
	"testing"
)

func TestValueFromGoValue(t *testing.T) {
	type foo struct {
		A string
		B map[string]int
	}
	f := foo{A: "hello", B: map[string]int{"a": 1, "b": 2}}
	v := ValueFromGoValue(f)
	log.Printf("v.typ=%s, kind=%s", v.typ(), v.Kind())
	log.Printf("v=%s", v)
	dyVal := v.(dynamicValue)
	log.Printf("dyVal.dy type=%T", dyVal.dy)
	stObj := dyVal.dy.(StructObject)
	log.Printf("a=%+v", stObj.GetField("A"))
	log.Printf("b=%+v", stObj.GetField("B"))
}

func TestReflection(t *testing.T) {
	type foo struct {
		a string
		b map[string]int
	}
	f := foo{a: "hello", b: map[string]int{"a": 1, "b": 2}}
	var g any = f
	ty := reflect.TypeOf(g)
	k := ty.Kind()
	t.Logf("ty=%+v, k=%s", ty, k)
	n := ty.NumField()
	t.Logf("numField=%d", n)
	v := reflect.ValueOf(f)
	for i := 0; i < n; i++ {
		t.Logf("i=%d, name=%s, val=%+v", i, ty.Field(i).Name, v.Field(i))
	}
}
