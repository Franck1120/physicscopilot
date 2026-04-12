import 'dart:async';

import 'package:camera/camera.dart';
import 'package:flutter/foundation.dart';
import 'package:image/image.dart' as img;

// ── Isolate types ─────────────────────────────────────────────────────────────

/// Result returned by the background isolate — processed JPEG + average luminance.
typedef _FrameResult = ({Uint8List? jpeg, double luminance});

/// Top-level function executed in an isolate via [compute].
/// Decodes [bytes] as JPEG, resizes to 512×512, re-encodes at quality 60.
/// Also computes average luminance (0–255) by sampling every 32nd pixel.
_FrameResult _processFrameIsolate(Uint8List bytes) {
  final decoded = img.decodeImage(bytes);
  if (decoded == null) return (jpeg: null, luminance: 128.0);

  final resized = img.copyResize(decoded, width: 512, height: 512);

  // Compute average luminance from a coarse pixel sample (16×16 grid).
  var lum = 0.0;
  var count = 0;
  for (var y = 0; y < resized.height; y += 32) {
    for (var x = 0; x < resized.width; x += 32) {
      final p = resized.getPixel(x, y);
      lum += 0.299 * p.r + 0.587 * p.g + 0.114 * p.b;
      count++;
    }
  }

  return (
    jpeg: img.encodeJpg(resized, quality: 60),
    luminance: count > 0 ? lum / count : 128.0,
  );
}

// ── FrameQuality ──────────────────────────────────────────────────────────────

/// Perceptual brightness classification of the last captured frame.
enum FrameQuality {
  /// Luminance in a normal range — no warning needed.
  ok,

  /// Average luminance < 40 — image too dark to analyse reliably.
  tooDark,

  /// Average luminance > 215 — image too bright / overexposed.
  tooBright,
}

FrameQuality _classify(double luminance) {
  if (luminance < 40) return FrameQuality.tooDark;
  if (luminance > 215) return FrameQuality.tooBright;
  return FrameQuality.ok;
}

// ── CameraService ─────────────────────────────────────────────────────────────

/// Manages the device camera lifecycle and produces a stream of compressed
/// frames suitable for transmission to the backend.
///
/// Frame rate is adaptive: starts at ~3 fps and backs off to ~2 fps when
/// the device cannot keep up. Identical frames (detected via a fast hash)
/// are skipped before encoding.
class CameraService {
  CameraController? _controller;
  Timer? _captureTimer;
  bool _isBusy = false;
  int? _lastHash;
  Duration _captureInterval = const Duration(milliseconds: 333); // ~3 fps

  final _frameController = StreamController<Uint8List>.broadcast();
  final _qualityController = StreamController<FrameQuality>.broadcast();

  /// Emits compressed 512×512 JPEG frames ready for transmission.
  Stream<Uint8List> get frames => _frameController.stream;

  /// Emits the brightness classification of every processed frame.
  Stream<FrameQuality> get quality => _qualityController.stream;

  /// The underlying [CameraController]; available after [initialize].
  CameraController? get controller => _controller;

  bool get isInitialized => _controller?.value.isInitialized ?? false;

  /// Initialises the rear camera and starts the periodic capture loop.
  Future<void> initialize() async {
    final cameras = await availableCameras();
    if (cameras.isEmpty) throw Exception('Nessuna camera disponibile');

    final rear = cameras.firstWhere(
      (c) => c.lensDirection == CameraLensDirection.back,
      orElse: () => cameras.first,
    );

    _controller = CameraController(
      rear,
      ResolutionPreset.medium,
      enableAudio: false,
      imageFormatGroup: ImageFormatGroup.jpeg,
    );

    await _controller!.initialize();
    _startCapturing();
  }

  void _startCapturing() {
    _captureTimer = Timer.periodic(_captureInterval, (_) => _captureFrame());
  }

  /// Captures a single frame immediately and returns the processed bytes.
  /// Returns null when the camera is busy or not initialised.
  Future<Uint8List?> captureFrame() async {
    if (_isBusy || !isInitialized) return null;
    _isBusy = true;
    try {
      final xFile = await _controller!.takePicture();
      final rawBytes = await xFile.readAsBytes();
      final result = await compute(_processFrameIsolate, rawBytes);
      if (!_qualityController.isClosed) {
        _qualityController.add(_classify(result.luminance));
      }
      return result.jpeg;
    } catch (_) {
      return null;
    } finally {
      _isBusy = false;
    }
  }

  Future<void> _captureFrame() async {
    if (_isBusy || !isInitialized) return;
    _isBusy = true;

    try {
      final sw = Stopwatch()..start();
      final xFile = await _controller!.takePicture();
      final rawBytes = await xFile.readAsBytes();
      sw.stop();

      // Adaptive back-off: if capture took >200 ms, reduce to ~2 fps.
      if (sw.elapsedMilliseconds > 200 &&
          _captureInterval.inMilliseconds < 400) {
        _captureTimer?.cancel();
        _captureInterval = const Duration(milliseconds: 500);
        _captureTimer =
            Timer.periodic(_captureInterval, (_) => _captureFrame());
      }

      // Motion detection: skip frames whose hash matches the previous one.
      final hash = _simpleHash(rawBytes);
      if (hash == _lastHash) return;
      _lastHash = hash;

      // Resize, re-encode and compute luminance in a background isolate.
      final result = await compute(_processFrameIsolate, rawBytes);

      if (!_qualityController.isClosed) {
        _qualityController.add(_classify(result.luminance));
      }
      if (result.jpeg != null && !_frameController.isClosed) {
        _frameController.add(result.jpeg!);
      }
    } catch (_) {
      // Silently skip individual frame failures.
    } finally {
      _isBusy = false;
    }
  }

  /// Fast rolling hash over 64 sampled bytes — good enough for motion detection.
  int _simpleHash(Uint8List bytes) {
    if (bytes.isEmpty) return 0;
    const sampleCount = 64;
    final step = bytes.length ~/ sampleCount;
    if (step == 0) return bytes.length;
    var hash = 0;
    for (var i = 0; i < bytes.length; i += step) {
      hash = (hash * 31 + bytes[i]) & 0xFFFFFFFF;
    }
    return hash;
  }

  Future<void> dispose() async {
    _captureTimer?.cancel();
    await _frameController.close();
    await _qualityController.close();
    await _controller?.dispose();
    _controller = null;
  }
}
