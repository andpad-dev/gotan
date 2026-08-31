package inlineexample

import "testing"

func TestHello(t *testing.T)              { _ = Hello() }
func BenchmarkHello(b *testing.B)         { _ = Hello() }
func BenchHello(b *testing.B)             { _ = Hello() }
func ExampleHello()                       { _ = Hello() }
func TestSomethingElse(t *testing.T)      { _ = Hello() }
func TestHello_withSuffix(t *testing.T)   { _ = Hello() }
func BenchmarkSomethingElse(b *testing.B) { _ = Hello() }
