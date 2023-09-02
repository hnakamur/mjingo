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
			input: f64Value{F: math.NaN()},
			want:  "NaN",
		},
		{
			input: f64Value{F: math.Inf(-1)},
			want:  "-inf",
		},
		{
			input: f64Value{F: math.Inf(+1)},
			want:  "inf",
		},
		{
			input: f64Value{F: 3.0},
			want:  "3.0",
		},
		{
			input: f64Value{F: 3.1},
			want:  "3.1",
		},
		{
			input: f64Value{F: float64(1e99)},
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
