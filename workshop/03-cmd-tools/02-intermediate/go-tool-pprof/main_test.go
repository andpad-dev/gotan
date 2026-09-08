package main

import "testing"

func TestTransform(t *testing.T) {
	if got, want := transform(1), uint64(15662720274509501185); got != want {
		t.Fatalf("transform(1) = %d, want %d", got, want)
	}
}

func TestManyTransforms(t *testing.T) {
	for i := uint64(0); i < 100; i++ {
		result ^= transform(i)
	}
}

func BenchmarkTransform(b *testing.B) {
	for b.Loop() {
		result = transform(1)
	}
}
