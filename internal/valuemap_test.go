package internal

import (
	"testing"
)

func TestValueMap(t *testing.T) {
	t.Run("strKeyRef", func(t *testing.T) {
		m := NewValueMap()

		m.Set(KeyRefFromString("foo"), Undefined)
		{
			got, ok := m.Get(KeyRefFromString("foo"))
			if !ok {
				t.Error("load ok mismatch")
			}
			if got != Undefined {
				t.Error("value mismatch")
			}
		}
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		m.Set(StrKeyRef{str: "bar"}, i64Value{n: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := m.EntryAt(uint(0))
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
			got, ok := m.EntryAt(uint(1))
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

		m.Set(StrKeyRef{str: "foo"}, f64Value{f: 3.1})
		{
			got, ok := m.EntryAt(uint(0))
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

		m.Delete(StrKeyRef{str: "foo"})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := m.EntryAt(uint(0))
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
		m := NewValueMap()

		m.Set(valueKeyRef{val: f64Value{f: 0.5}}, Undefined)
		{
			got, ok := m.Get(valueKeyRef{val: f64Value{f: 0.5}})
			if !ok {
				t.Error("load ok mismatch")
			}
			if got != Undefined {
				t.Error("value mismatch")
			}
		}
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		m.Set(valueKeyRef{val: i64Value{n: 123}}, i64Value{n: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := m.EntryAt(uint(0))
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
			got, ok := m.EntryAt(uint(1))
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

		m.Set(valueKeyRef{val: f64Value{f: 0.5}}, f64Value{f: 3.1})
		{
			got, ok := m.EntryAt(uint(0))
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

		m.Delete(valueKeyRef{val: f64Value{f: 0.5}})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := m.EntryAt(uint(0))
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
		m := NewValueMap()

		m.Set(valueKeyRef{val: f64Value{f: 0.5}}, Undefined)
		{
			got, ok := m.Get(valueKeyRef{val: f64Value{f: 0.5}})
			if !ok {
				t.Error("load ok mismatch")
			}
			if got != Undefined {
				t.Error("value mismatch")
			}
		}
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		m.Set(StrKeyRef{str: "bar"}, i64Value{n: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := m.EntryAt(uint(0))
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
			got, ok := m.EntryAt(uint(1))
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

		m.Set(valueKeyRef{val: f64Value{f: 0.5}}, f64Value{f: 3.1})
		{
			got, ok := m.EntryAt(uint(0))
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

		m.Delete(valueKeyRef{val: f64Value{f: 0.5}})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := m.EntryAt(uint(0))
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
