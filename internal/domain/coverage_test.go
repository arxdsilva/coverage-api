package domain

import "testing"

func TestCompareCoverage(t *testing.T) {
	t.Run("new value", func(t *testing.T) {
		delta, direction := CompareCoverage(80.0, nil)
		if delta != nil {
			t.Fatalf("expected nil delta")
		}
		if direction != DirectionNew {
			t.Fatalf("expected direction new, got %s", direction)
		}
	})

	t.Run("up", func(t *testing.T) {
		prev := 79.0
		delta, direction := CompareCoverage(80.5, &prev)
		if delta == nil || *delta != 1.5 {
			t.Fatalf("expected delta 1.5, got %v", delta)
		}
		if direction != DirectionUp {
			t.Fatalf("expected direction up, got %s", direction)
		}
	})

	t.Run("down", func(t *testing.T) {
		prev := 81.0
		delta, direction := CompareCoverage(80.5, &prev)
		if delta == nil || *delta != -0.5 {
			t.Fatalf("expected delta -0.5, got %v", delta)
		}
		if direction != DirectionDown {
			t.Fatalf("expected direction down, got %s", direction)
		}
	})

	t.Run("equal", func(t *testing.T) {
		prev := 80.5
		delta, direction := CompareCoverage(80.5, &prev)
		if delta == nil || *delta != 0 {
			t.Fatalf("expected delta 0, got %v", delta)
		}
		if direction != DirectionEqual {
			t.Fatalf("expected direction equal, got %s", direction)
		}
	})
}

func TestEvaluateThreshold(t *testing.T) {
	if got := EvaluateThreshold(80.0, 80.0); got != ThresholdPassed {
		t.Fatalf("expected passed, got %s", got)
	}
	if got := EvaluateThreshold(79.9, 80.0); got != ThresholdFailed {
		t.Fatalf("expected failed, got %s", got)
	}
}
