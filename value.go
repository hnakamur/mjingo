package mjingo

import "github.com/hnakamur/mjingo/internal/value"

type Value = value.Value
type ValueFromGoValueOption = value.ValueFromGoValueOption

var Undefined = value.Undefined

func ValueFromSafeString(s string) Value {
	return value.ValueFromSafeString(s)
}

func WithStructTag(tag string) ValueFromGoValueOption {
	return value.WithStructTag(tag)
}

func ValueFromGoValue(val any, opts ...ValueFromGoValueOption) Value {
	return value.ValueFromGoValue(val, opts...)
}
