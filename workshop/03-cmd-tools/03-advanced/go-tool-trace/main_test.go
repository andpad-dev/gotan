package main

import (
	"context"
	"runtime/trace"
	"testing"
)

func TestProcessWork(t *testing.T) {
	ctx, task := trace.NewTask(context.Background(), "work")
	defer task.End()

	processWork(ctx)
}
