import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/providers/camera_provider.dart';
import 'package:physicscopilot/services/camera_service.dart';

// ── Fake CameraService ────────────────────────────────────────────────────────

class FakeCameraService extends CameraService {
  bool initialized = false;
  bool disposed = false;

  @override
  Future<void> initialize() async => initialized = true;

  @override
  Future<void> dispose() async => disposed = true;

  @override
  Stream<FrameQuality> get quality => const Stream.empty();
}

// ── Tests ─────────────────────────────────────────────────────────────────────

void main() {
  group('cameraServiceProvider', () {
    test('returns a CameraService instance', () {
      final fake = FakeCameraService();
      final container = ProviderContainer(
        overrides: [cameraServiceProvider.overrideWithValue(fake)],
      );
      addTearDown(container.dispose);

      final service = container.read(cameraServiceProvider);
      expect(service, isA<CameraService>());
      expect(service, same(fake));
    });
  });

  group('cameraInitProvider', () {
    test('is a FutureProvider<void> — compile-time check via AsyncValue', () async {
      final fake = FakeCameraService();
      final container = ProviderContainer(
        overrides: [cameraServiceProvider.overrideWithValue(fake)],
      );
      addTearDown(container.dispose);

      final result = container.read(cameraInitProvider);
      // FutureProvider returns AsyncValue — verify the type signature at runtime
      expect(result, isA<AsyncValue<void>>());
    });

    test('calling initialize() on FakeCameraService sets initialized=true', () async {
      final fake = FakeCameraService();
      final container = ProviderContainer(
        overrides: [cameraServiceProvider.overrideWithValue(fake)],
      );
      addTearDown(container.dispose);

      // Reading cameraInitProvider triggers initialize() on the service
      await container.read(cameraInitProvider.future);
      expect(fake.initialized, isTrue);
    });
  });

  group('frameQualityProvider', () {
    test('is a StreamProvider<FrameQuality> — compile-time check via AsyncValue', () {
      final fake = FakeCameraService();
      final container = ProviderContainer(
        overrides: [cameraServiceProvider.overrideWithValue(fake)],
      );
      addTearDown(container.dispose);

      final result = container.read(frameQualityProvider);
      // StreamProvider exposes AsyncValue — verify the type signature at runtime
      expect(result, isA<AsyncValue<FrameQuality>>());
    });
  });

  group('Container disposal', () {
    test('container.dispose() does not throw', () {
      final fake = FakeCameraService();
      final container = ProviderContainer(
        overrides: [cameraServiceProvider.overrideWithValue(fake)],
      );

      // Read the provider so it is actually created and its onDispose is registered
      container.read(cameraServiceProvider);

      expect(() => container.dispose(), returnsNormally);
    });
  });
}
