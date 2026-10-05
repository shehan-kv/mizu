package project

import "testing"

func TestNewMetric(t *testing.T) {
	metric := NewMetric("tasks_completed", 42)

	if metric.Key() != "tasks_completed" {
		t.Fatalf(
			"expected key %q, got %q",
			"tasks_completed",
			metric.Key(),
		)
	}

	if metric.Value() != 42 {
		t.Fatalf(
			"expected value %d, got %d",
			42,
			metric.Value(),
		)
	}
}
