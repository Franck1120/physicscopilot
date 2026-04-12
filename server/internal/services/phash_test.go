package services

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

// createTestJPEGBase64 generates a minimal JPEG image encoded as base64.
func createTestJPEGBase64(width, height int) string {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 3), G: uint8(y * 3), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75})
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

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

func TestPHashFrameValidJPEG(t *testing.T) {
	b64 := createTestJPEGBase64(64, 64)
	hash, err := pHashFrame(b64)
	if err != nil {
		t.Fatalf("unexpected error for valid JPEG: %v", err)
	}
	// The hash should be deterministic for the same image
	hash2, err := pHashFrame(b64)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if hash != hash2 {
		t.Errorf("expected same hash for same image, got %x vs %x", hash, hash2)
	}
}

func TestPHashFrameSimilarImagesCloseHamming(t *testing.T) {
	b64a := createTestJPEGBase64(64, 64)
	hashA, err := pHashFrame(b64a)
	if err != nil {
		t.Fatalf("image A: %v", err)
	}

	// Create a very similar image (slight color shift)
	imgB := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			imgB.Set(x, y, color.RGBA{R: uint8(x*3 + 1), G: uint8(y*3 + 1), B: 129, A: 255})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, imgB, &jpeg.Options{Quality: 75})
	b64b := base64.StdEncoding.EncodeToString(buf.Bytes())

	hashB, err := pHashFrame(b64b)
	if err != nil {
		t.Fatalf("image B: %v", err)
	}

	dist := hammingDistance(hashA, hashB)
	// Similar images should have low hamming distance
	if dist > 20 {
		t.Errorf("expected similar images to have low hamming distance, got %d", dist)
	}
}

func TestResizeToGrayscale(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.White)
		}
	}

	pixels := resizeToGrayscale(img, 8)
	if len(pixels) != 64 {
		t.Fatalf("expected 64 pixels (8x8), got %d", len(pixels))
	}
	// All-white image should produce all values close to 1.0
	for i, p := range pixels {
		if p < 0.99 || p > 1.01 {
			t.Errorf("pixel %d: expected ~1.0 for white image, got %f", i, p)
			break
		}
	}
}

func TestComputeDCT2D(t *testing.T) {
	// 4x4 all-zero pixels: DCT should be all zeros
	pixels := make([]float64, 16)
	dct := computeDCT2D(pixels, 4)
	if len(dct) != 4 {
		t.Fatalf("expected 4 rows, got %d", len(dct))
	}
	for u := 0; u < 4; u++ {
		for v := 0; v < 4; v++ {
			if dct[u][v] != 0.0 {
				t.Errorf("expected DCT[%d][%d]=0 for zero input, got %f", u, v, dct[u][v])
			}
		}
	}
}

func TestComputeDCT2DConstantInput(t *testing.T) {
	// Constant input: only DC coefficient (dct[0][0]) should be non-zero
	size := 8
	pixels := make([]float64, size*size)
	for i := range pixels {
		pixels[i] = 0.5
	}
	dct := computeDCT2D(pixels, size)
	if dct[0][0] == 0 {
		t.Error("expected non-zero DC coefficient for constant input")
	}
	// All AC coefficients should be ~0
	for u := 0; u < size; u++ {
		for v := 0; v < size; v++ {
			if u == 0 && v == 0 {
				continue
			}
			if dct[u][v] > 1e-10 || dct[u][v] < -1e-10 {
				t.Errorf("expected AC coefficient DCT[%d][%d] ~0, got %f", u, v, dct[u][v])
			}
		}
	}
}
