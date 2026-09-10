package main

import "testing"

func TestCalculate(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{input: 0, want: 0},
		{input: 3, want: 30},
	}

	for _, tt := range tests {
		if got := calculate(tt.input); got != tt.want {
			t.Fatalf("calculate(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}
