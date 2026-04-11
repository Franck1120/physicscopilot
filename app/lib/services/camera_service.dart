import 'dart:async';

import 'package:camera/camera.dart';
import 'package:flutter/foundation.dart';
import 'package:image/image.dart' as img;

/// Top-level function executed in an isolate via [compute].
/// Decodes [bytes] as JPEG, resizes to 512×512, re-encodes at quality 60.
Uint8List? _processFrameIsolate(Uint8List bytes) {
  final decoded = img.decodeImage(bytes);
  if (decoded == null) return null;
  final resized = img.copyResize(decoded, width: 512, height: 512);
  return img.encodeJpg(resized, quality: 60);
}

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

  /// Emits compressed 512×512 JPEG frames ready for transmission.
  Stream<Uint8List> get frames => _frameController.stream;

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
    _captureTimer =
        Timer.periodic(_captureInterval, (_) => _captureFrame());
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

      // Resize and re-encode in a background isolate.
      final processed = await compute(_processFrameIsolate, rawBytes);
      if (processed != null && !_frameController.isClosed) {
        _frameController.add(processed);
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
    await _controller?.dispose();
    _controller = null;
  }
}
