// Golden tests for HomeScreen.
// Run with: flutter test test/golden/home_screen_golden_test.dart --update-goldens
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
import 'package:physicscopilot/screens/home_screen.dart';
import 'package:physicscopilot/services/api_service.dart'
    show ServerHealth, serverHealthProvider;
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

void _noop() {}

void main() {
  setUpAll(() {
    GoogleFonts.config.allowRuntimeFetching = false;
  });

  group('HomeScreen golden', () {
    late SharedPreferences prefs;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      prefs = await SharedPreferences.getInstance();
    });

    testWidgets('HomeScreen — disconnected / empty sessions', (tester) async {
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
            serverHealthProvider.overrideWith(
              (ref) => Stream.value(ServerHealth.offline()),
            ),
          ],
          child: MaterialApp(
            debugShowCheckedModeBanner: false,
            home: HomeScreen(
              onChangeEquipment: _noop,
              onStartCamera: _noop,
            ),
          ),
        ),
      );

      statusController.add(ConnectionStatus.disconnected);
      await tester.pump();
      // Advance past the 400 ms skeleton-reveal timer in
      // _RecentSessionsSectionState.initState.
      await tester.pump(const Duration(milliseconds: 500));

      await expectLater(
        find.byType(MaterialApp),
        matchesGoldenFile('goldens/home_screen.png'),
      );
    });
  });
}
