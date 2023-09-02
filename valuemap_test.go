package mjingo

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

		m.Set(strKeyRef{str: "bar"}, I64Value{N: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "foo"}
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
			wantKey := strKeyRef{str: "bar"}
			if got.Key != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := I64Value{N: 3}
			if got.Value != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Set(strKeyRef{str: "foo"}, F64Value{F: 3.1})
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "foo"}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := F64Value{F: 3.1}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Delete(strKeyRef{str: "foo"})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "bar"}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := I64Value{N: 3}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
	})
	t.Run("ValueKeyRef", func(t *testing.T) {
		m := NewValueMap()

		m.Set(valueKeyRef{val: F64Value{F: 0.5}}, Undefined)
		{
			got, ok := m.Get(valueKeyRef{val: F64Value{F: 0.5}})
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

		m.Set(valueKeyRef{val: I64Value{N: 123}}, I64Value{N: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: F64Value{F: 0.5}}
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
			wantKey := valueKeyRef{val: I64Value{N: 123}}
			if got.Key != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := I64Value{N: 3}
			if got.Value != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Set(valueKeyRef{val: F64Value{F: 0.5}}, F64Value{F: 3.1})
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: F64Value{F: 0.5}}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := F64Value{F: 3.1}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Delete(valueKeyRef{val: F64Value{F: 0.5}})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: I64Value{N: 123}}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := I64Value{N: 3}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
	})
	t.Run("mixOfStrAndValueKeyRef", func(t *testing.T) {
		m := NewValueMap()

		m.Set(valueKeyRef{val: F64Value{F: 0.5}}, Undefined)
		{
			got, ok := m.Get(valueKeyRef{val: F64Value{F: 0.5}})
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

		m.Set(strKeyRef{str: "bar"}, I64Value{N: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: F64Value{F: 0.5}}
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
			wantKey := strKeyRef{str: "bar"}
			if got.Key != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := I64Value{N: 3}
			if got.Value != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Set(valueKeyRef{val: F64Value{F: 0.5}}, F64Value{F: 3.1})
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: F64Value{F: 0.5}}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := F64Value{F: 3.1}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Delete(valueKeyRef{val: F64Value{F: 0.5}})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "bar"}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := I64Value{N: 3}
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
	})
}
