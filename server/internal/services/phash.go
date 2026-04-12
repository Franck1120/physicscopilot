// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

// phash.go implements DCT-based perceptual hashing (pHash) for camera frames.
// Perceptual hashing lets ConversationService skip duplicate or near-identical
// frames without calling the Gemini API, reducing cost and latency.

package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg" // register JPEG decoder
	"math"
	"math/bits"
)

const (
	// pHashGridSize is the size of the grid used for DCT computation.
	pHashGridSize = 32
	// pHashDCTSize is the number of low-frequency DCT coefficients per axis.
	pHashDCTSize = 8
	// PHashDuplicateThreshold is the maximum Hamming distance for two frames to
	// be considered perceptually identical.
	PHashDuplicateThreshold = 5
)

// pHashFrame computes a DCT-based perceptual hash (pHash) of a base64-encoded
// JPEG frame. Returns an error if the data cannot be decoded as JPEG.
// Only standard and URL-safe base64 encodings are tried.
func pHashFrame(frameBase64 string) (uint64, error) {
	data, err := base64.StdEncoding.DecodeString(frameBase64)
	if err != nil {
		data, err = base64.URLEncoding.DecodeString(frameBase64)
		if err != nil {
			return 0, fmt.Errorf("decode base64: %w", err)
		}
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("decode image: %w", err)
	}

	pixels := resizeToGrayscale(img, pHashGridSize)
	dct := computeDCT2D(pixels, pHashGridSize)

	// Flatten top-left 8x8 DCT block (low-frequency components).
	ac := make([]float64, pHashDCTSize*pHashDCTSize)
	for v := 0; v < pHashDCTSize; v++ {
		for u := 0; u < pHashDCTSize; u++ {
			ac[v*pHashDCTSize+u] = dct[v][u]
		}
	}

	// Mean of AC values (skip DC at index 0 to improve robustness).
	var sum float64
	for _, v := range ac[1:] {
		sum += v
	}
	mean := sum / float64(len(ac)-1)

	// Build 64-bit hash: bit i = 1 if ac[i] > mean (skip i=0, DC component).
	var hash uint64
	for i, v := range ac {
		if i > 0 && v > mean {
			hash |= 1 << uint(i)
		}
	}
	return hash, nil
}

// hammingDistance returns the number of differing bits between two 64-bit hashes.
func hammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

// resizeToGrayscale samples img at a size x size grid and returns the normalised
// grayscale intensities (0.0-1.0) in row-major order.
func resizeToGrayscale(img image.Image, size int) []float64 {
	bounds := img.Bounds()
	srcW := bounds.Max.X - bounds.Min.X
	srcH := bounds.Max.Y - bounds.Min.Y

	pixels := make([]float64, size*size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			srcX := bounds.Min.X + x*srcW/size
			srcY := bounds.Min.Y + y*srcH/size
			r, g, b, _ := img.At(srcX, srcY).RGBA()
			// Rec. 601 luma: Y = 0.299R + 0.587G + 0.114B (values are 16-bit).
			gray := (299.0*float64(r) + 587.0*float64(g) + 114.0*float64(b)) / (1000.0 * 65535.0)
			pixels[y*size+x] = gray
		}
	}
	return pixels
}

// computeDCT2D computes the 2-D DCT-II of a size x size pixel array.
// Returns a 2-D slice of DCT coefficients.
func computeDCT2D(pixels []float64, size int) [][]float64 {
	n := float64(size)
	dct := make([][]float64, size)
	for i := range dct {
		dct[i] = make([]float64, size)
	}
	for u := 0; u < size; u++ {
		cu := 1.0
		if u == 0 {
			cu = 1.0 / math.Sqrt2
		}
		for v := 0; v < size; v++ {
			cv := 1.0
			if v == 0 {
				cv = 1.0 / math.Sqrt2
			}
			var sum float64
			for x := 0; x < size; x++ {
				cosU := math.Cos(math.Pi * float64(2*x+1) * float64(u) / (2 * n))
				for y := 0; y < size; y++ {
					cosV := math.Cos(math.Pi * float64(2*y+1) * float64(v) / (2 * n))
					sum += pixels[x*size+y] * cosU * cosV
				}
			}
			dct[u][v] = (2.0 / n) * cu * cv * sum
		}
	}
	return dct
}
