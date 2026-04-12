package services

import (
	"math"
	"testing"
)

// TestPHashFrameNilBytesReturnsError verifies that passing a nil (empty)
// string to pHashFrame returns an error rather than panicking.
// Go has no nil string — the equivalent is an empty string "".
// pHashFrame("") must return a non-nil error because "" is not valid base64
// that decodes to a JPEG.
func TestPHashFrameNilBytesReturnsError(t *testing.T) {
	t.Parallel()

	_, err := pHashFrame("")
	if err == nil {
		t.Error("expected error for empty string (nil-equivalent) input, got nil")
	}
}

// TestPHashFrameEmptySliceBase64ReturnsError verifies that base64-encoded
// empty bytes (the empty string) returns an error.
// base64.StdEncoding.EncodeToString([]byte{}) == "" so this is equivalent.
func TestPHashFrameEmptySliceBase64ReturnsError(t *testing.T) {
	t.Parallel()

	// An empty byte slice encodes to "" in base64.
	_, err := pHashFrame("")
	if err == nil {
		t.Error("expected error for base64 of empty slice, got nil")
	}
}

// TestPHashFrameNonImageDataReturnsError verifies that valid base64 that does
// not decode to a JPEG (plain ASCII "hello world") returns an error from the
// image-decode stage.
func TestPHashFrameNonImageDataReturnsError(t *testing.T) {
	t.Parallel()

	// "aGVsbG8gd29ybGQ=" is base64 for "hello world" — valid base64, not JPEG.
	_, err := pHashFrame("aGVsbG8gd29ybGQ=")
	if err == nil {
		t.Error("expected error for non-JPEG base64 data, got nil")
	}
}

// TestHammingDistanceZeroZeroIsZero verifies the identity property:
// hammingDistance(0, 0) == 0.
func TestHammingDistanceZeroZeroIsZero(t *testing.T) {
	t.Parallel()

	if d := hammingDistance(0, 0); d != 0 {
		t.Errorf("hammingDistance(0, 0) = %d, want 0", d)
	}
}

// TestHammingDistanceMaxUint64And0Is64 verifies that comparing math.MaxUint64
// (all 64 bits set) against 0 (no bits set) yields 64 differing bits.
func TestHammingDistanceMaxUint64And0Is64(t *testing.T) {
	t.Parallel()

	d := hammingDistance(math.MaxUint64, 0)
	if d != 64 {
		t.Errorf("hammingDistance(MaxUint64, 0) = %d, want 64", d)
	}
}

// TestHammingDistanceBothMaxUint64IsZero verifies that two identical maximum
// values have a Hamming distance of zero.
func TestHammingDistanceBothMaxUint64IsZero(t *testing.T) {
	t.Parallel()

	d := hammingDistance(math.MaxUint64, math.MaxUint64)
	if d != 0 {
		t.Errorf("hammingDistance(MaxUint64, MaxUint64) = %d, want 0", d)
	}
}

// TestPHashFrameInvalidBase64NoPanic verifies that malformed base64 returns
// an error without panicking.
func TestPHashFrameInvalidBase64NoPanic(t *testing.T) {
	t.Parallel()

	inputs := []string{
		"!!! not base64 !!!",
		"@@@@@@",
		"====",
		"\x00\x01\x02",
	}
	for _, input := range inputs {
		_, err := pHashFrame(input)
		if err == nil {
			t.Errorf("pHashFrame(%q): expected error for invalid input, got nil", input)
		}
	}
}
