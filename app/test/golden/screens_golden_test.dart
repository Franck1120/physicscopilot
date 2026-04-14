// ignore_for_file: avoid_redundant_argument_values
//
// Golden tests — pixel-level regression snapshots for key app screens.
//
// To create or update golden files run:
//   flutter test test/golden/screens_golden_test.dart --update-goldens
//
// Golden files are stored in test/golden/goldens/ and must be committed.
// These snapshots are platform-specific; re-generate on the target CI OS.

import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart' show onboardingCompletedProvider;
import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/providers/websocket_provider.dart';
import 'package:physicscopilot/screens/home_screen.dart';
import 'package:physicscopilot/screens/onboarding_screen.dart';
import 'package:physicscopilot/services/api_service.dart' show serverHealthProvider, ServerHealth;
import 'package:physicscopilot/services/websocket_service.dart';

// Fixed viewport mimicking an iPhone 14 logical pixel footprint.
const _kViewSize = Size(390.0, 844.0);

void main() {
  setUpAll(() {
    // Prevent google_fonts from attempting network fetches inside tests.
    GoogleFonts.config.allowRuntimeFetching = false;
  });

  group('golden screens', () {
    late SharedPreferences prefs;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      prefs = await SharedPreferences.getInstance();
    });

    testWidgets('OnboardingScreen — first slide', (tester) async {
      _fixViewport(tester);

      await tester.pumpWidget(
        ProviderScope(
          // Start with onboarding not completed so the screen renders fully.
          overrides: [
            sharedPrefsProvider.overrideWithValue(prefs),
            onboardingCompletedProvider.overrideWith((ref) => false),
          ],
          child: const MaterialApp(
            debugShowCheckedModeBanner: false,
            home: OnboardingScreen(onComplete: _noop),
          ),
        ),
      );

      // Let layout settle; skip animations that never complete (e.g. cursor blink).
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 50));

      await expectLater(
        find.byType(MaterialApp),
        matchesGoldenFile('goldens/onboarding_screen_first_slide.png'),
      );
    });

    testWidgets('HomeScreen — empty sessions state', (tester) async {
      _fixViewport(tester);

      final statusController = StreamController<ConnectionStatus>.broadcast();
      addTearDown(statusController.close);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            sharedPrefsProvider.overrideWithValue(prefs),
            // Provide a disconnected status so the connection banner shows
            // without any real network activity.
            webSocketServiceProvider.overrideWithValue(
              WebSocketService('ws://localhost:0'),
            ),
            connectionStatusProvider.overrideWith(
              (ref) => statusController.stream,
            ),
            // Override serverHealthProvider to avoid the 15-second polling
            // loop that would leave a pending timer after the test.
            serverHealthProvider.overrideWith(
              (ref) => Stream.value(ServerHealth.offline()),
            ),
          ],
          child: const MaterialApp(
            debugShowCheckedModeBanner: false,
            home: HomeScreen(
              onChangeEquipment: _noop,
              onStartCamera: _noop,
            ),
          ),
        ),
      );

      // Emit disconnected so the UI is in a deterministic state.
      statusController.add(ConnectionStatus.disconnected);
      await tester.pump();
      // Advance the clock past the 400 ms skeleton-reveal timer in
      // _RecentSessionsSectionState.initState so no timers remain pending.
      await tester.pump(const Duration(milliseconds: 500));

      await expectLater(
        find.byType(MaterialApp),
        matchesGoldenFile('goldens/home_screen_empty_state.png'),
      );
    });
  });
}

void _fixViewport(WidgetTester tester) {
  tester.view.physicalSize = _kViewSize * tester.view.devicePixelRatio;
  tester.view.devicePixelRatio = 2.0;
  addTearDown(() {
    tester.view.resetPhysicalSize();
    tester.view.resetDevicePixelRatio();
  });
}

void _noop() {}
