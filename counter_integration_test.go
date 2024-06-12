//go:build integration
// +build integration

package main

import (
	"context"
	"testing"
)

func TestIncreaseCounter(t *testing.T) {
	ctx := context.Background()
	inc := Counter(NewDB())

	prev, err := inc(ctx)
	if err != nil {
		t.Error(err)
	}

	next, err := inc(ctx)
	if err != nil {
		t.Error(err)
	}

	if prev+1 != next {
		t.Errorf("want counter %d increased by one, got incr %d", next, next-prev)
	}
}
