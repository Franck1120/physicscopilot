// Golden tests for CameraScreen.
// Run with: flutter test test/golden/camera_screen_golden_test.dart --update-goldens
//
// CameraScreen requires a live camera hardware channel.  In golden tests the
// camera provider is kept perpetually in loading state (Completer never
// completes) so the loading/initialising UI is captured without a real device.
//
// Golden files are stored in test/golden/goldens/ and must be committed.
// These snapshots are platform-specific; re-generate on the target CI OS.

import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/providers/camera_provider.dart';
import 'package:physicscopilot/providers/websocket_provider.dart';
import 'package:physicscopilot/screens/camera_screen.dart';
import 'package:physicscopilot/services/camera_service.dart';
import 'package:physicscopilot/services/websocket_service.dart';

// ---------------------------------------------------------------------------
// Fake CameraService that never resolves so the provider stays AsyncLoading.
// ---------------------------------------------------------------------------

class _NeverInitCameraService extends CameraService {
  @override
  Future<void> initialize() async {
    // Block forever — keeps cameraInitProvider in AsyncLoading state.
    await Completer<void>().future;
  }

  @override
  Stream<FrameQuality> get quality => const Stream.empty();
}

const _kViewSize = Size(390.0, 844.0);

void _fixViewport(WidgetTester tester) {
  tester.view.physicalSize = _kViewSize * tester.view.devicePixelRatio;
  tester.view.devicePixelRatio = 2.0;
  addTearDown(() {
    tester.view.resetPhysicalSize();
    tester.view.resetDevicePixelRatio();
  });
}

void main() {
  setUpAll(() {
    GoogleFonts.config.allowRuntimeFetching = false;
  });

  group('CameraScreen golden', () {
    late SharedPreferences prefs;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      prefs = await SharedPreferences.getInstance();
    });

    testWidgets('CameraScreen — camera loading state', (tester) async {
      _fixViewport(tester);

      final statusController = StreamController<ConnectionStatus>.broadcast();
      addTearDown(statusController.close);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            sharedPrefsProvider.overrideWithValue(prefs),
            cameraServiceProvider.overrideWithValue(_NeverInitCameraService()),
            webSocketServiceProvider.overrideWithValue(
              WebSocketService('ws://localhost:0'),
            ),
            connectionStatusProvider.overrideWith(
              (ref) => statusController.stream,
            ),
          ],
          child: const MaterialApp(
            debugShowCheckedModeBanner: false,
            home: CameraScreen(),
          ),
        ),
      );

      statusController.add(ConnectionStatus.disconnected);
      // Single pump — captures loading UI before camera platform channel fires.
      await tester.pump();

      await expectLater(
        find.byType(MaterialApp),
        matchesGoldenFile('goldens/camera_screen.png'),
      );
    });
  });
}
