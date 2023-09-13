package mjingo

import (
	"testing"
)

func TestValueMap(t *testing.T) {
	t.Run("strKeyRef", func(t *testing.T) {
		m := newValueMap()

		m.Set(keyRefFromString("foo"), Undefined)
		{
			got, ok := m.Get(keyRefFromString("foo"))
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

		m.Set(strKeyRef{str: "bar"}, valueFromI64(3))
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
			wantVal := valueFromI64(3)
			if got.Value != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Set(strKeyRef{str: "foo"}, valueFromF64(3.1))
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "foo"}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := valueFromF64(3.1)
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
			wantVal := valueFromI64(3)
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
	})
	t.Run("ValueKeyRef", func(t *testing.T) {
		m := newValueMap()

		m.Set(valueKeyRef{val: valueFromF64(0.5)}, Undefined)
		{
			got, ok := m.Get(valueKeyRef{val: valueFromF64(0.5)})
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

		m.Set(valueKeyRef{val: valueFromI64(123)}, valueFromI64(3))
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: valueFromF64(0.5)}
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
			wantKey := valueKeyRef{val: valueFromI64(123)}
			if got.Key != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := valueFromI64(3)
			if got.Value != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Set(valueKeyRef{val: valueFromF64(0.5)}, valueFromF64(3.1))
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: valueFromF64(0.5)}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := valueFromF64(3.1)
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Delete(valueKeyRef{val: valueFromF64(0.5)})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: valueFromI64(123)}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := valueFromI64(3)
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
	})
	t.Run("mixOfStrAndValueKeyRef", func(t *testing.T) {
		m := newValueMap()

		m.Set(valueKeyRef{val: valueFromF64(0.5)}, Undefined)
		{
			got, ok := m.Get(valueKeyRef{val: valueFromF64(0.5)})
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

		m.Set(strKeyRef{str: "bar"}, valueFromI64(3))
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: valueFromF64(0.5)}
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
			wantVal := valueFromI64(3)
			if got.Value != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Set(valueKeyRef{val: valueFromF64(0.5)}, valueFromF64(3.1))
		{
			got, ok := m.EntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: valueFromF64(0.5)}
			if got.Key != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", got.Key, wantKey)
			}
			wantVal := valueFromF64(3.1)
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}

		m.Delete(valueKeyRef{val: valueFromF64(0.5)})
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
			wantVal := valueFromI64(3)
			if got.Value != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", got.Value, wantVal)
			}
		}
	})
}
