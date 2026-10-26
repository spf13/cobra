package cobra

import (
	"testing"
)

func TestEqMixedKindsDoNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Eq panicked on mixed supported kinds: %v", r)
		}
	}()
	if Eq(1, "1") {
		t.Fatalf("expected false for mixed kinds, got true")
	}
	if !Eq(1, 1) {
		t.Fatalf("expected true for equal ints")
	}
	if !Eq("a", "a") {
		t.Fatalf("expected true for equal strings")
	}
}
