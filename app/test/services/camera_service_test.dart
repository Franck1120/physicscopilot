// Unit tests for CameraService.
//
// The camera hardware is unavailable in test, so we verify:
// 1. Initial state (not initialised, streams are live).
// 2. captureFrame() returns null when camera is not initialised.
// 3. dispose() closes streams and cleans up.
// 4. FrameQuality enum values cover all brightness classifications.
// 5. initialize() throws when no cameras are available.
import 'dart:async';

import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/services/camera_service.dart';

void main() {
  group('CameraService — initial state', () {
    late CameraService service;

    setUp(() {
      service = CameraService();
    });

    tearDown(() async {
      await service.dispose();
    });

    test('isInitialized is false before initialize()', () {
      expect(service.isInitialized, isFalse);
    });

    test('controller is null before initialize()', () {
      expect(service.controller, isNull);
    });

    test('frames stream is a broadcast stream', () {
      expect(service.frames.isBroadcast, isTrue);
    });

    test('quality stream is a broadcast stream', () {
      expect(service.quality.isBroadcast, isTrue);
    });
  });

  group('CameraService — captureFrame() guard clauses', () {
    late CameraService service;

    setUp(() {
      service = CameraService();
    });

    tearDown(() async {
      await service.dispose();
    });

    test('captureFrame() returns null when camera is not initialised',
        () async {
      final result = await service.captureFrame();
      expect(result, isNull);
    });

    test(
        'captureFrame() can be called multiple times safely '
        'when not initialised', () async {
      final results = await Future.wait([
        service.captureFrame(),
        service.captureFrame(),
        service.captureFrame(),
      ]);
      expect(results, everyElement(isNull));
    });
  });

  group('CameraService — dispose()', () {
    test('dispose() completes without error on fresh instance', () async {
      final service = CameraService();
      await expectLater(service.dispose(), completes);
    });

    test('dispose() closes the frames stream', () async {
      final service = CameraService();
      final framesSub = service.frames.listen((_) {});
      await service.dispose();

      // After dispose, the stream subscription should have received a done event.
      // Verify by checking that adding a new listener immediately gets done.
      var doneReceived = false;
      service.frames.listen(
        (_) {},
        onDone: () => doneReceived = true,
      );

      // Give the microtask queue a chance to deliver the done event.
      await Future<void>.delayed(Duration.zero);
      expect(doneReceived, isTrue);

      await framesSub.cancel();
    });

    test('dispose() closes the quality stream', () async {
      final service = CameraService();
      final qualitySub = service.quality.listen((_) {});
      await service.dispose();

      var doneReceived = false;
      service.quality.listen(
        (_) {},
        onDone: () => doneReceived = true,
      );

      await Future<void>.delayed(Duration.zero);
      expect(doneReceived, isTrue);

      await qualitySub.cancel();
    });

    test('dispose() sets controller to null', () async {
      final service = CameraService();
      await service.dispose();
      expect(service.controller, isNull);
    });

    test('captureFrame() returns null after dispose()', () async {
      final service = CameraService();
      await service.dispose();
      final result = await service.captureFrame();
      expect(result, isNull);
    });
  });

  group('CameraService — initialize() error handling', () {
    test('initialize() throws when no cameras are available', () async {
      // In test environment availableCameras() throws a MissingPluginException
      // because no camera plugin is registered. This verifies that
      // CameraService.initialize() propagates the error rather than
      // silently swallowing it.
      final service = CameraService();
      addTearDown(service.dispose);

      expect(
        () => service.initialize(),
        throwsA(anything),
      );
    });
  });

  group('FrameQuality — enum completeness', () {
    test('FrameQuality has exactly three values', () {
      expect(FrameQuality.values, hasLength(3));
    });

    test('FrameQuality contains ok, tooDark, and tooBright', () {
      expect(
        FrameQuality.values,
        containsAll([FrameQuality.ok, FrameQuality.tooDark, FrameQuality.tooBright]),
      );
    });

    test('FrameQuality.ok is distinct from error states', () {
      expect(FrameQuality.ok, isNot(FrameQuality.tooDark));
      expect(FrameQuality.ok, isNot(FrameQuality.tooBright));
    });

    test('FrameQuality values have stable indices for serialisation safety',
        () {
      // Guard against accidental reordering that could break persisted data.
      expect(FrameQuality.ok.index, equals(0));
      expect(FrameQuality.tooDark.index, equals(1));
      expect(FrameQuality.tooBright.index, equals(2));
    });
  });
}
