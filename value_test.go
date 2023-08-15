package mjingo

import (
	"math"
	"testing"
)

func TestValueString(t *testing.T) {
	testCases := []struct {
		input value
		want  string
	}{
		{
			input: value{typ: valueTypeF64, data: math.NaN()},
			want:  "NaN",
		},
		{
			input: value{typ: valueTypeF64, data: math.Inf(-1)},
			want:  "-inf",
		},
		{
			input: value{typ: valueTypeF64, data: math.Inf(+1)},
			want:  "inf",
		},
		{
			input: value{typ: valueTypeF64, data: 3.0},
			want:  "3.0",
		},
		{
			input: value{typ: valueTypeF64, data: 3.1},
			want:  "3.1",
		},
		{
			input: value{typ: valueTypeF64, data: float64(1e99)},
			want:  "1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000.0",
		},
	}
	for _, tc := range testCases {
		got := tc.input.String()
		if got != tc.want {
			t.Errorf("result mismatch, input=%+v, got=%s, want=%s", tc.input, got, tc.want)
		}
	}
}
