package services

import (
	"testing"
)

func TestHammingDistanceSameHash(t *testing.T) {
	if d := hammingDistance(0xDEADBEEF, 0xDEADBEEF); d != 0 {
		t.Errorf("same hash: want distance 0, got %d", d)
	}
}

func TestHammingDistanceAllDifferent(t *testing.T) {
	d := hammingDistance(0, 0xFFFFFFFFFFFFFFFF)
	if d != 64 {
		t.Errorf("all bits different: want 64, got %d", d)
	}
}

func TestHammingDistanceOneBitDiff(t *testing.T) {
	d := hammingDistance(0b01, 0b11)
	if d != 1 {
		t.Errorf("one bit diff: want 1, got %d", d)
	}
}

func TestPHashFrameNonJPEGReturnsError(t *testing.T) {
	_, err := pHashFrame("not-base64-jpeg")
	if err == nil {
		t.Error("expected error for non-JPEG input")
	}
}
