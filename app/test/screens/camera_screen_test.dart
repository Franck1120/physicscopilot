// Widget tests for CameraScreen.
//
// CameraScreen is a full-screen camera view with WebSocket bridging.
// Tests avoid instantiating a real CameraController (requires device camera)
// by keeping cameraInitProvider in the loading or error state.
// Connection-status and session state are overridden via ProviderScope.
import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/providers/camera_provider.dart';
import 'package:physicscopilot/providers/settings_provider.dart';
import 'package:physicscopilot/providers/websocket_provider.dart';
import 'package:physicscopilot/providers/voice_provider.dart';
import 'package:physicscopilot/services/websocket_service.dart';
import 'package:physicscopilot/services/camera_service.dart';
import 'package:physicscopilot/services/voice_service.dart';
import 'package:physicscopilot/screens/camera_screen.dart';
import 'package:physicscopilot/main.dart' show sharedPrefsProvider;

// ── Fake CameraService that never initialises (stays loading) ─────────────────

class _NeverInitCameraService extends CameraService {
  @override
  Future<void> initialize() async {
    // Block forever — keeps cameraInitProvider in AsyncLoading.
    await Completer<void>().future;
  }

  @override
  Stream<FrameQuality> get quality => const Stream.empty();
}

// ── Fake CameraService that immediately throws ────────────────────────────────

class _ErrorCameraService extends CameraService {
  @override
  Future<void> initialize() async {
    throw Exception('Camera non disponibile in test');
  }

  @override
  Stream<FrameQuality> get quality => const Stream.empty();
}

// ── Helper: shared prefs + base provider overrides ───────────────────────────

Future<List<Override>> _baseOverrides({
  required CameraService cameraService,
  Stream<ConnectionStatus>? statusStream,
}) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();

  return [
    sharedPrefsProvider.overrideWithValue(prefs),
    settingsProvider.overrideWith((ref) => SettingsNotifier(prefs)),
    cameraServiceProvider.overrideWithValue(cameraService),
    webSocketServiceProvider.overrideWith((ref) {
      final svc = WebSocketService('ws://localhost:19999');
      ref.onDispose(() => svc.disconnect());
      return svc;
    }),
    connectionStatusProvider.overrideWith(
      (ref) =>
          statusStream ?? Stream.value(ConnectionStatus.connected),
    ),
    voiceServiceProvider.overrideWithValue(VoiceService()),
  ];
}

// ── Tests ─────────────────────────────────────────────────────────────────────

void main() {
  group('CameraScreen', () {
    testWidgets('renders without crash (loading state)', (tester) async {
      final overrides = await _baseOverrides(
        cameraService: _NeverInitCameraService(),
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: overrides,
          child: const MaterialApp(home: CameraScreen()),
        ),
      );
      await tester.pump();

      // Screen mounts without exception.
      expect(find.byType(CameraScreen), findsOneWidget);
    });

    testWidgets('shows CircularProgressIndicator while camera is loading',
        (tester) async {
      final overrides = await _baseOverrides(
        cameraService: _NeverInitCameraService(),
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: overrides,
          child: const MaterialApp(home: CameraScreen()),
        ),
      );
      await tester.pump();

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows error text when camera initialisation fails',
        (tester) async {
      final overrides = await _baseOverrides(
        cameraService: _ErrorCameraService(),
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: overrides,
          child: const MaterialApp(home: CameraScreen()),
        ),
      );
      // Let the Future resolve and rebuild.
      await tester.pumpAndSettle();

      expect(find.textContaining('Errore camera'), findsOneWidget);
    });

    testWidgets(
        'no offline banner shown when connectionStatus is connected (loading camera)',
        (tester) async {
      final overrides = await _baseOverrides(
        cameraService: _NeverInitCameraService(),
        statusStream: Stream.value(ConnectionStatus.connected),
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: overrides,
          child: const MaterialApp(home: CameraScreen()),
        ),
      );
      await tester.pump(const Duration(milliseconds: 100));

      // _OfflineBanner uses AnimatedSlide/AnimatedOpacity — when isVisible is
      // false the banner is slid out of view.  We verify no red error banner
      // text is rendered (the banner container is still in the tree but hidden).
      expect(find.byIcon(Icons.wifi_off), findsNothing);
    });

    testWidgets(
        'offline banner icon visible when connectionStatus is disconnected',
        (tester) async {
      final overrides = await _baseOverrides(
        cameraService: _NeverInitCameraService(),
        statusStream: Stream.value(ConnectionStatus.disconnected),
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: overrides,
          child: const MaterialApp(home: CameraScreen()),
        ),
      );
      // Allow StreamProvider to emit the disconnected value and let the
      // AnimatedSlide/AnimatedOpacity start their animations.
      await tester.pump(const Duration(milliseconds: 100));
      // Drive animations to completion.
      await tester.pumpAndSettle();

      // The _OfflineBanner renders a wifi_off icon when isVisible == true.
      expect(find.byIcon(Icons.wifi_off), findsOneWidget);
    });

    testWidgets('Scaffold background is black', (tester) async {
      final overrides = await _baseOverrides(
        cameraService: _NeverInitCameraService(),
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: overrides,
          child: const MaterialApp(home: CameraScreen()),
        ),
      );
      await tester.pump();

      final scaffold = tester.widget<Scaffold>(find.byType(Scaffold));
      expect(scaffold.backgroundColor, equals(Colors.black));
    });
  });
}
