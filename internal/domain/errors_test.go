package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrSourceMismatch(t *testing.T) {
	t.Run("ErrSourceMismatch is defined", func(t *testing.T) {
		if ErrSourceMismatch == nil {
			t.Fatal("ErrSourceMismatch should not be nil")
		}
	})

	t.Run("ErrSourceMismatch message", func(t *testing.T) {
		want := "skill source type does not match filter"
		if ErrSourceMismatch.Error() != want {
			t.Errorf("ErrSourceMismatch.Error() = %q, want %q", ErrSourceMismatch.Error(), want)
		}
	})

	t.Run("errors.Is works with ErrSourceMismatch", func(t *testing.T) {
		wrapped := errors.New("wrapped: " + ErrSourceMismatch.Error())
		if errors.Is(wrapped, ErrSourceMismatch) {
			t.Error("errors.Is should not match a non-wrapped error")
		}

		wrapped2 := fmt.Errorf("context: %w", ErrSourceMismatch)
		if !errors.Is(wrapped2, ErrSourceMismatch) {
			t.Error("errors.Is should match when ErrSourceMismatch is wrapped with %w")
		}
	})
}
