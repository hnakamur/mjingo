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
			input: f64Value{f: math.NaN()},
			want:  "NaN",
		},
		{
			input: f64Value{f: math.Inf(-1)},
			want:  "-inf",
		},
		{
			input: f64Value{f: math.Inf(+1)},
			want:  "inf",
		},
		{
			input: f64Value{f: 3.0},
			want:  "3.0",
		},
		{
			input: f64Value{f: 3.1},
			want:  "3.1",
		},
		{
			input: f64Value{f: float64(1e99)},
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
