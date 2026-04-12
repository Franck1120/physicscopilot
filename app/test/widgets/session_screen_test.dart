// Widget tests for SessionScreen.
//
// Camera and WebSocket providers are overridden with no-op fakes so the tests
// run without any real hardware or network.  The focus is on the connection
// banner that surfaces when the WebSocket is not in the connected state.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/providers/camera_provider.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/providers/websocket_provider.dart';
import 'package:physicscopilot/screens/session_screen.dart';
import 'package:physicscopilot/services/camera_service.dart';
import 'package:physicscopilot/services/websocket_service.dart';
import 'package:physicscopilot/widgets/connection_banner.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Builds a [ProviderScope] with [SessionScreen] and all dangerous providers
/// replaced by in-process fakes.
///
/// [connectionStatus] controls the value that [connectionStatusProvider] emits.
Widget _buildSession({
  required ConnectionStatus connectionStatus,
  required SharedPreferences prefs,
}) {
  return ProviderScope(
    overrides: [
      // SharedPreferences — avoids reading real device storage.
      sharedPrefsProvider.overrideWithValue(prefs),

      // WebSocketService — returns valid broadcast streams; never connects.
      webSocketServiceProvider.overrideWithValue(
        WebSocketService('ws://localhost:0'),
      ),

      // CameraService — safe to construct without platform channels.
      cameraServiceProvider.overrideWithValue(CameraService()),

      // cameraInitProvider — completes immediately so CameraPreviewWidget
      // gets a data state without triggering availableCameras() platform call.
      cameraInitProvider.overrideWith((_) async {}),

      // connectionStatusProvider — controlled stream for test scenarios.
      connectionStatusProvider.overrideWith(
        (_) => Stream.value(connectionStatus),
      ),
    ],
    child: const MaterialApp(home: SessionScreen()),
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('SessionScreen — connection banner visibility', () {
    testWidgets('shows banner with disconnected message when offline', (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      await tester.pumpWidget(_buildSession(
        connectionStatus: ConnectionStatus.disconnected,
        prefs: prefs,
      ));
      // First pump builds the widget tree; second pump triggers post-frame callbacks.
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      expect(find.byType(ConnectionBanner), findsOneWidget);
      expect(
        find.text('Server non raggiungibile — i dati non vengono inviati'),
        findsOneWidget,
      );
    });

    testWidgets('shows banner with connecting message while reconnecting', (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      await tester.pumpWidget(_buildSession(
        connectionStatus: ConnectionStatus.connecting,
        prefs: prefs,
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      expect(find.byType(ConnectionBanner), findsOneWidget);
      expect(find.text('Connessione al server in corso…'), findsOneWidget);
    });

    testWidgets('hides banner when WebSocket is connected', (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      await tester.pumpWidget(_buildSession(
        connectionStatus: ConnectionStatus.connected,
        prefs: prefs,
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      expect(find.byType(ConnectionBanner), findsNothing);
    });
  });

  group('SessionScreen — structural elements', () {
    testWidgets('renders AppBar with "Sessione attiva" title', (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      await tester.pumpWidget(_buildSession(
        connectionStatus: ConnectionStatus.connected,
        prefs: prefs,
      ));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      expect(find.text('Sessione attiva'), findsOneWidget);
    });

    testWidgets('has a back button in the AppBar', (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      await tester.pumpWidget(_buildSession(
        connectionStatus: ConnectionStatus.connected,
        prefs: prefs,
      ));
      await tester.pump();

      expect(find.byIcon(Icons.arrow_back_ios_new), findsOneWidget);
    });
  });
}
