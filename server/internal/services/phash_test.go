package services

import (
	"image"
	"image/color"
	"math"
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

// ── resizeToGrayscale ─────────────────────────────────────────────────────────

func TestResizeToGrayscaleWhiteImage(t *testing.T) {
	// 4x4 solid white image → all pixels should be ~1.0.
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}

	pixels := resizeToGrayscale(img, 4)
	if len(pixels) != 16 {
		t.Fatalf("expected 16 pixels for 4x4 grid, got %d", len(pixels))
	}
	for i, p := range pixels {
		if p < 0.99 {
			t.Errorf("pixel[%d]: expected ~1.0 for white, got %f", i, p)
		}
	}
}

func TestResizeToGrayscaleBlackImage(t *testing.T) {
	// 4x4 solid black image → all pixels should be 0.0.
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	// NewRGBA initialises to transparent (all zeros) — already black.

	pixels := resizeToGrayscale(img, 4)
	for i, p := range pixels {
		if p != 0.0 {
			t.Errorf("pixel[%d]: expected 0.0 for black, got %f", i, p)
		}
	}
}

func TestResizeToGrayscaleOutputSize(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 100, 100))
	pixels := resizeToGrayscale(img, pHashGridSize)
	if len(pixels) != pHashGridSize*pHashGridSize {
		t.Errorf("expected %d pixels, got %d", pHashGridSize*pHashGridSize, len(pixels))
	}
}

// ── computeDCT2D ──────────────────────────────────────────────────────────────

func TestComputeDCT2DOutputShape(t *testing.T) {
	size := 4
	pixels := make([]float64, size*size)
	dct := computeDCT2D(pixels, size)
	if len(dct) != size {
		t.Fatalf("expected %d rows, got %d", size, len(dct))
	}
	for i, row := range dct {
		if len(row) != size {
			t.Errorf("row[%d]: expected %d cols, got %d", i, size, len(row))
		}
	}
}

func TestComputeDCT2DConstantInputDCNonZero(t *testing.T) {
	// A constant non-zero signal → only the DC component (dct[0][0]) is non-zero.
	size := 4
	pixels := make([]float64, size*size)
	for i := range pixels {
		pixels[i] = 0.5
	}
	dct := computeDCT2D(pixels, size)

	if dct[0][0] == 0 {
		t.Error("DC component must be non-zero for a constant non-zero input")
	}
	// All AC components should be near zero.
	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			if u == 0 && v == 0 {
				continue
			}
			if math.Abs(dct[u][v]) > 1e-9 {
				t.Errorf("AC component dct[%d][%d] should be ~0 for constant input, got %f", u, v, dct[u][v])
			}
		}
	}
}

func TestComputeDCT2DZeroInputAllZero(t *testing.T) {
	size := 4
	pixels := make([]float64, size*size) // all zeros
	dct := computeDCT2D(pixels, size)
	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			if dct[u][v] != 0 {
				t.Errorf("dct[%d][%d]: expected 0 for zero input, got %f", u, v, dct[u][v])
			}
		}
	}
}
