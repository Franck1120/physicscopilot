// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

package services

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

// makeWhiteJPEGBase64 encodes an 8×8 all-white JPEG at quality 90 and returns
// the standard base64 string. Called once per benchmark sub-test for setup.
func makeWhiteJPEGBase64(b *testing.B) string {
	b.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, white)
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		b.Fatalf("jpeg.Encode: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// BenchmarkPHashFrame measures the full pHashFrame pipeline: base64 decode,
// JPEG decode, grayscale resize, 2-D DCT, and hash construction.
func BenchmarkPHashFrame(b *testing.B) {
	b.ReportAllocs()

	jpegB64 := makeWhiteJPEGBase64(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := pHashFrame(jpegB64); err != nil {
			b.Fatalf("pHashFrame: %v", err)
		}
	}
}

// BenchmarkHammingDistance measures the O(1) Hamming-distance computation.
// This benchmark demonstrates the bitcount cost and acts as a baseline for
// callers that compare many hashes in a tight loop.
func BenchmarkHammingDistance(b *testing.B) {
	b.ReportAllocs()

	const a uint64 = 0xDEADBEEF_CAFEBABE
	const c uint64 = 0xCAFEBABE_DEADBEEF

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hammingDistance(a, c)
	}
}

// BenchmarkPHashFrameParallel is the concurrent variant of BenchmarkPHashFrame.
// It verifies that pHashFrame is safe to call from multiple goroutines
// (each with its own stack — no shared mutable state) and measures throughput
// under parallelism.
func BenchmarkPHashFrameParallel(b *testing.B) {
	b.ReportAllocs()

	jpegB64 := makeWhiteJPEGBase64(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := pHashFrame(jpegB64); err != nil {
				b.Errorf("pHashFrame: %v", err)
			}
		}
	})
}
