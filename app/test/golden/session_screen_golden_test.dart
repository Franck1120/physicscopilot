// Golden tests for SessionScreen.
// Run with: flutter test test/golden/session_screen_golden_test.dart --update-goldens
//
// SessionScreen requires camera and WebSocket providers. In golden tests the
// camera provider is kept in the loading state (never initialises) so the
// initial skeleton/loading UI is captured without device hardware.
//
// Golden files are stored in test/golden/goldens/ and must be committed.
// These snapshots are platform-specific; re-generate on the target CI OS.

import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart' show sharedPrefsProvider;
import 'package:physicscopilot/providers/websocket_provider.dart';
import 'package:physicscopilot/screens/session_screen.dart';
import 'package:physicscopilot/services/websocket_service.dart';

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

  group('SessionScreen golden', () {
    late SharedPreferences prefs;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      prefs = await SharedPreferences.getInstance();
    });

    testWidgets('SessionScreen — loading / disconnected state', (tester) async {
      _fixViewport(tester);

      final statusController = StreamController<ConnectionStatus>.broadcast();
      addTearDown(statusController.close);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            sharedPrefsProvider.overrideWithValue(prefs),
            webSocketServiceProvider.overrideWithValue(
              WebSocketService('ws://localhost:0'),
            ),
            connectionStatusProvider.overrideWith(
              (ref) => statusController.stream,
            ),
          ],
          child: const MaterialApp(
            debugShowCheckedModeBanner: false,
            home: SessionScreen(),
          ),
        ),
      );

      statusController.add(ConnectionStatus.disconnected);
      // Single pump — captures the initial loading/skeleton state before
      // the camera platform channel is invoked.
      await tester.pump();

      await expectLater(
        find.byType(MaterialApp),
        matchesGoldenFile('goldens/session_screen.png'),
      );
    });
  });
}
