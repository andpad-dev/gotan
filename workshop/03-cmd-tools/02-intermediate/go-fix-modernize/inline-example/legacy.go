package inlineexample

//go:fix inline
func Hello() string { return "hello" }

func UseHello() string { return Hello() }
