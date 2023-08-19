package value

import (
	"testing"
)

func TestValueIndexMap(t *testing.T) {
	t.Run("strKeyRef", func(t *testing.T) {
		m := newValueIndexMap()

		m.Store(StrKeyRef{str: "foo"}, Undefined)
		{
			got, ok := m.Load(StrKeyRef{str: "foo"})
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

		m.Store(StrKeyRef{str: "bar"}, i64Value{n: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "foo"}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := Undefined
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(1))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "bar"}
			if gotKey != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := i64Value{n: 3}
			if gotVal != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Store(StrKeyRef{str: "foo"}, f64Value{f: 3.1})
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "foo"}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := f64Value{f: 3.1}
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Delete(StrKeyRef{str: "foo"})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "bar"}
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

		m.Store(ValueKeyRef{val: f64Value{f: 0.5}}, Undefined)
		{
			got, ok := m.Load(ValueKeyRef{val: f64Value{f: 0.5}})
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

		m.Store(ValueKeyRef{val: i64Value{n: 123}}, i64Value{n: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := ValueKeyRef{val: f64Value{f: 0.5}}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := Undefined
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(1))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := ValueKeyRef{val: i64Value{n: 123}}
			if gotKey != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := i64Value{n: 3}
			if gotVal != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Store(ValueKeyRef{val: f64Value{f: 0.5}}, f64Value{f: 3.1})
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := ValueKeyRef{val: f64Value{f: 0.5}}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := f64Value{f: 3.1}
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Delete(ValueKeyRef{val: f64Value{f: 0.5}})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := ValueKeyRef{val: i64Value{n: 123}}
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

		m.Store(ValueKeyRef{val: f64Value{f: 0.5}}, Undefined)
		{
			got, ok := m.Load(ValueKeyRef{val: f64Value{f: 0.5}})
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

		m.Store(StrKeyRef{str: "bar"}, i64Value{n: 3})
		if got, want := m.Len(), uint(2); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}

		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := ValueKeyRef{val: f64Value{f: 0.5}}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := Undefined
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(1))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "bar"}
			if gotKey != wantKey {
				t.Errorf("second key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := i64Value{n: 3}
			if gotVal != wantVal {
				t.Errorf("second value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Store(ValueKeyRef{val: f64Value{f: 0.5}}, f64Value{f: 3.1})
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := ValueKeyRef{val: f64Value{f: 0.5}}
			if gotKey != wantKey {
				t.Errorf("first key mismatch, got=%+v, want=%+v", gotKey, wantKey)
			}
			wantVal := f64Value{f: 3.1}
			if gotVal != wantVal {
				t.Errorf("first value mismatch, got=%+v, want=%+v", gotVal, wantVal)
			}
		}

		m.Delete(ValueKeyRef{val: f64Value{f: 0.5}})
		if got, want := m.Len(), uint(1); got != want {
			t.Errorf("len mismatch, got=%d, want=%d", got, want)
		}
		{
			gotKey, gotVal, ok := m.GetEntryAt(uint(0))
			if !ok {
				t.Error("load ok mismatch")
			}
			wantKey := StrKeyRef{str: "bar"}
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
