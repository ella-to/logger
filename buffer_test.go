package logger

import (
	"testing"
	"time"
)

func TestBuffer(t *testing.T) {
	var result []int

	buffer := newBuffer[int](10, 1*time.Second, func(records []int) error {
		result = append(result, records...)
		return nil
	})

	for i := 0; i < 10; i++ {
		buffer.Add(i)
	}

	if len(result) > 0 {
		t.Errorf("expected 0, got %d", len(result))
	}

	time.Sleep(2 * time.Second)

	if len(result) != 10 {
		t.Errorf("expected 10, got %d", len(result))
	}

	for i := 0; i < 10; i++ {
		buffer.Add(i)
	}

	time.Sleep(2 * time.Second)

	if len(result) != 20 {
		t.Errorf("expected 20, got %d", len(result))
	}
}
