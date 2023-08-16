package mjingo

import (
	"testing"
)

func TestValueIndexMap(t *testing.T) {
	t.Run("strKeyRef", func(t *testing.T) {
		m := newValueIndexMap()

		m.Store(strKeyRef{str: "foo"}, valueUndefined)
		{
			got, ok := m.Load(strKeyRef{str: "foo"})
			if !ok {
				t.Error("load ok mismatch")
			}
			if got != valueUndefined {
				t.Error("value mismatch")
			}
		}
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		m.Store(strKeyRef{str: "bar"}, i64Value{n: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "foo"}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := valueUndefined
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(1))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "bar"}
			if gotKey != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := i64Value{n: 3}
			if gotVal != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Store(strKeyRef{str: "foo"}, f64Value{f: 3.1})
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "foo"}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := f64Value{f: 3.1}
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Delete(strKeyRef{str: "foo"})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "bar"}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := i64Value{n: 3}
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}
	})
	t.Run("ValueKeyRef", func(t *testing.T) {
		m := newValueIndexMap()

		m.Store(valueKeyRef{val: f64Value{f: 0.5}}, valueUndefined)
		{
			got, ok := m.Load(valueKeyRef{val: f64Value{f: 0.5}})
			if !ok {
				t.Error("load ok mismatch")
			}
			if got != valueUndefined {
				t.Error("value mismatch")
			}
		}
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		m.Store(valueKeyRef{val: i64Value{n: 123}}, i64Value{n: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: f64Value{f: 0.5}}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := valueUndefined
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(1))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: i64Value{n: 123}}
			if gotKey != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := i64Value{n: 3}
			if gotVal != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Store(valueKeyRef{val: f64Value{f: 0.5}}, f64Value{f: 3.1})
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: f64Value{f: 0.5}}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := f64Value{f: 3.1}
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Delete(valueKeyRef{val: f64Value{f: 0.5}})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: i64Value{n: 123}}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := i64Value{n: 3}
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}
	})
	t.Run("mixOfStrAndValueKeyRef", func(t *testing.T) {
		m := newValueIndexMap()

		m.Store(valueKeyRef{val: f64Value{f: 0.5}}, valueUndefined)
		{
			got, ok := m.Load(valueKeyRef{val: f64Value{f: 0.5}})
			if !ok {
				t.Error("load ok mismatch")
			}
			if got != valueUndefined {
				t.Error("value mismatch")
			}
		}
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		m.Store(strKeyRef{str: "bar"}, i64Value{n: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: f64Value{f: 0.5}}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := valueUndefined
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(1))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "bar"}
			if gotKey != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := i64Value{n: 3}
			if gotVal != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Store(valueKeyRef{val: f64Value{f: 0.5}}, f64Value{f: 3.1})
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := valueKeyRef{val: f64Value{f: 0.5}}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := f64Value{f: 3.1}
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Delete(valueKeyRef{val: f64Value{f: 0.5}})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := strKeyRef{str: "bar"}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := i64Value{n: 3}
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}
	})
}
