package mjingo

import (
	"math"
	"testing"
)

func TestValueString(t *testing.T) {
	testCases := []struct {
		input Value
		want  string
	}{
		{
			input: valueFromF64(math.NaN()),
			want:  "NaN",
		},
		{
			input: valueFromF64(math.Inf(-1)),
			want:  "-inf",
		},
		{
			input: valueFromF64(math.Inf(+1)),
			want:  "inf",
		},
		{
			input: valueFromF64(3.0),
			want:  "3.0",
		},
		{
			input: valueFromF64(3.1),
			want:  "3.1",
		},
		{
			input: valueFromF64(float64(1e99)),
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
