// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

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

func TestHammingDistanceHalfBitsDifferent(t *testing.T) {
	// Lower 32 bits set vs upper 32 bits set — all 64 bits differ.
	d := hammingDistance(0x00000000FFFFFFFF, 0xFFFFFFFF00000000)
	if d != 64 {
		t.Errorf("non-overlapping halves: want distance 64, got %d", d)
	}
}

func TestHammingDistanceSymmetric(t *testing.T) {
	a := uint64(0xCAFEBABE12345678)
	b := uint64(0xDEADBEEF87654321)
	if hammingDistance(a, b) != hammingDistance(b, a) {
		t.Errorf("hammingDistance must be symmetric: hammingDistance(%x, %x) != hammingDistance(%x, %x)",
			a, b, b, a)
	}
}

func TestPHashFrameEmptyStringReturnsError(t *testing.T) {
	_, err := pHashFrame("")
	if err == nil {
		t.Error("expected error for empty string input")
	}
}

// minimalWhiteJPEG8x8 is a valid 8×8 all-white JPEG encoded in standard base64.
// Generated with Go's image/jpeg encoder at quality 90.
const minimalWhiteJPEG8x8 = "/9j/2wCEAAMCAgMCAgMDAwMEAwMEBQgFBQQEBQoHBwYIDAoMDAsKCwsNDhIQDQ4RDgsLEBYQERMUFRUVDA8XGBYUGBIUFRQBAwQEBQQFCQUFCRQNCw0UFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFP/AABEIAAgACAMBIgACEQEDEQH/xAGiAAABBQEBAQEBAQAAAAAAAAAAAQIDBAUGBwgJCgsQAAIBAwMCBAMFBQQEAAABfQECAwAEEQUSITFBBhNRYQcicRQygZGhCCNCscEVUtHwJDNicoIJChYXGBkaJSYnKCkqNDU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6g4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2drh4uPk5ebn6Onq8fLz9PX29/j5+gEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoLEQACAQIEBAMEBwUEBAABAncAAQIDEQQFITEGEkFRB2FxEyIygQgUQpGhscEJIzNS8BVictEKFiQ04SXxFxgZGiYnKCkqNTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqCg4SFhoeIiYqSk5SVlpeYmZqio6Slpqeoqaqys7S1tre4ubrCw8TFxsfIycrS09TV1tfY2dri4+Tl5ufo6ery8/T19vf4+fr/2gAMAwEAAhEDEQA/AP1TooooA//Z"

func TestPHashFrameValidMinimalJPEGReturnsHash(t *testing.T) {
	hash, err := pHashFrame(minimalWhiteJPEG8x8)
	if err != nil {
		t.Fatalf("expected no error for valid JPEG, got: %v", err)
	}
	// A uniform white image produces a deterministic hash — assert it's stable across calls.
	hash2, err := pHashFrame(minimalWhiteJPEG8x8)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if hash != hash2 {
		t.Errorf("pHashFrame must be deterministic: first=%x second=%x", hash, hash2)
	}
}

func TestPHashFrameInvalidBase64ReturnsError(t *testing.T) {
	_, err := pHashFrame("!!!not-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64 input")
	}
}
