package main_test

import (
	"testing"
	"time"
)

func TestAlpha(t *testing.T) {
	t.Parallel()
	t.Run("A", func(t *testing.T) {
		t.Parallel()
		time.Sleep(2 * time.Second)
	})
	t.Run("B", func(t *testing.T) {
		t.Parallel()
		time.Sleep(1 * time.Second)
		t.Run("X", func(t *testing.T) {
			t.Parallel()
		})
		time.Sleep(1 * time.Second)
	})
	t.Run("C", func(t *testing.T) {
		t.Parallel()
		time.Sleep(3 * time.Second)
	})
}

func TestBeta(t *testing.T) {
	t.Parallel()
	t.Run("A", func(t *testing.T) {
		t.Parallel()
		time.Sleep(1 * time.Second)
	})
	t.Run("B", func(t *testing.T) {
		t.Parallel()
		time.Sleep(2 * time.Second)
	})
	t.Run("C", func(t *testing.T) {
		t.Parallel()
		time.Sleep(3 * time.Second)
		t.Error("failure")
	})
}
