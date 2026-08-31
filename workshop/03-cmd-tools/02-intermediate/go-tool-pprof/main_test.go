package main

import "testing"

func TestSeatCode(t *testing.T) {
	if got, want := seatCode(1), uint64(15662720274509501185); got != want {
		t.Fatalf("seatCode(1) = %d, want %d", got, want)
	}
}

func TestIssueCodes(t *testing.T) {
	for i := uint64(0); i < 100; i++ {
		result ^= seatCode(i)
	}
}

func BenchmarkSeatCode(b *testing.B) {
	for b.Loop() {
		result = seatCode(1)
	}
}
