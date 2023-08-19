package value

import (
	"testing"
)

func TestValueIndexMap(t *testing.T) {
	t.Run("strKeyRef", func(t *testing.T) {
		m := NewIndexMap()

		IndexMapSet(m, KeyRefFromString("foo"), Undefined)
		{
			got, ok := IndexMapGet(m, KeyRefFromString("foo"))
			if !ok {
				t.Error("load ok mismatch")
			}
			if got != Undefined {
				t.Error("value mismatch")
			}
		}
		if got, want := IndexMapLen(m), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		IndexMapSet(m, StrKeyRef{str: "bar"}, i64Value{n: 3})
		if got, want := IndexMapLen(m), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := IndexMapEntryAt(m, uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "foo"}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := Undefined
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
		{
			got, ok := IndexMapEntryAt(m, uint(1))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "bar"}
			if got.Key != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := i64Value{n: 3}
			if got.Value != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		IndexMapSet(m, StrKeyRef{str: "foo"}, f64Value{f: 3.1})
		{
			got, ok := IndexMapEntryAt(m, uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "foo"}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := f64Value{f: 3.1}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		IndexMapDelete(m, StrKeyRef{str: "foo"})
		if got, want := IndexMapLen(m), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := IndexMapEntryAt(m, uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "bar"}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := i64Value{n: 3}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
	})
	t.Run("ValueKeyRef", func(t *testing.T) {
		m := NewIndexMap()

		IndexMapSet(m, valueKeyRef{val: f64Value{f: 0.5}}, Undefined)
		{
			got, ok := IndexMapGet(m, valueKeyRef{val: f64Value{f: 0.5}})
			if !ok {
				t.Error("load ok mismatch")
			}
			if got != Undefined {
				t.Error("value mismatch")
			}
		}
		if got, want := IndexMapLen(m), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		IndexMapSet(m, valueKeyRef{val: i64Value{n: 123}}, i64Value{n: 3})
		if got, want := IndexMapLen(m), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := IndexMapEntryAt(m, uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: f64Value{f: 0.5}}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := Undefined
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
		{
			got, ok := IndexMapEntryAt(m, uint(1))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: i64Value{n: 123}}
			if got.Key != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := i64Value{n: 3}
			if got.Value != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		IndexMapSet(m, valueKeyRef{val: f64Value{f: 0.5}}, f64Value{f: 3.1})
		{
			got, ok := IndexMapEntryAt(m, uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: f64Value{f: 0.5}}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := f64Value{f: 3.1}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		IndexMapDelete(m, valueKeyRef{val: f64Value{f: 0.5}})
		if got, want := IndexMapLen(m), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := IndexMapEntryAt(m, uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: i64Value{n: 123}}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := i64Value{n: 3}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
	})
	t.Run("mixOfStrAndValueKeyRef", func(t *testing.T) {
		m := NewIndexMap()

		IndexMapSet(m, valueKeyRef{val: f64Value{f: 0.5}}, Undefined)
		{
			got, ok := IndexMapGet(m, valueKeyRef{val: f64Value{f: 0.5}})
			if !ok {
				t.Error("load ok mismatch")
			}
			if got != Undefined {
				t.Error("value mismatch")
			}
		}
		if got, want := IndexMapLen(m), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		IndexMapSet(m, StrKeyRef{str: "bar"}, i64Value{n: 3})
		if got, want := IndexMapLen(m), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := IndexMapEntryAt(m, uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: f64Value{f: 0.5}}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := Undefined
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
		{
			got, ok := IndexMapEntryAt(m, uint(1))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "bar"}
			if got.Key != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := i64Value{n: 3}
			if got.Value != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		IndexMapSet(m, valueKeyRef{val: f64Value{f: 0.5}}, f64Value{f: 3.1})
		{
			got, ok := IndexMapEntryAt(m, uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: f64Value{f: 0.5}}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := f64Value{f: 3.1}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		IndexMapDelete(m, valueKeyRef{val: f64Value{f: 0.5}})
		if got, want := IndexMapLen(m), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := IndexMapEntryAt(m, uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "bar"}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := i64Value{n: 3}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
	})
}
