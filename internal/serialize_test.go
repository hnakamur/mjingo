package internal

import (
	"log"
	"testing"
)

func TestValueFromGoValue(t *testing.T) {
	type foo struct {
		A string `json:"a"`
		B map[string]int
	}
	f := foo{A: "hello", B: map[string]int{"a": 1, "b": 2}}
	v := ValueFromGoValue(f, WithStructTag("json"))
	log.Printf("v.typ=%s, kind=%s", v.typ(), v.Kind())
	log.Printf("v=%s", v)
	dyVal := v.(dynamicValue)
	log.Printf("dyVal.dy type=%T", dyVal.dy)
	stObj := dyVal.dy.(StructObject)
	log.Printf("a=%+v", stObj.GetField("a"))
	log.Printf("b=%+v", stObj.GetField("B"))
}

func TestValueFromGoValueLoop(t *testing.T) {
	type foo struct {
		a    string
		b    map[string]int
		Self *foo
	}
	f := foo{a: "hello", b: map[string]int{"a": 1, "b": 2}}
	f.Self = &f
	v := ValueFromGoValue(f, WithStructTag("json"))
	dyVal := v.(dynamicValue)
	stObj := dyVal.dy.(StructObject)
	selfVal := stObj.GetField("Self")
	log.Printf("self=%+v", selfVal)
	selfVal2 := selfVal.Unwrap().(dynamicValue).dy.(StructObject).GetField("Self")
	log.Printf("self2=%+v", selfVal2)
}