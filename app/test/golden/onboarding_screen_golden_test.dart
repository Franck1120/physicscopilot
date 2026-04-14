// Golden tests for OnboardingScreen.
// Run with: flutter test test/golden/onboarding_screen_golden_test.dart --update-goldens
//
// Golden files are stored in test/golden/goldens/ and must be committed.
// These snapshots are platform-specific; re-generate on the target CI OS.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart' show onboardingCompletedProvider;
import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/screens/onboarding_screen.dart';

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

  group('OnboardingScreen golden', () {
    late SharedPreferences prefs;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      prefs = await SharedPreferences.getInstance();
    });

    testWidgets('OnboardingScreen — first slide', (tester) async {
      _fixViewport(tester);

      await tester.pumpWidget(
        ProviderScope(
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
      await tester.pump();
      // Advance slightly to let layout animations settle while avoiding
      // indefinitely-pending timers (e.g. cursor blink).
      await tester.pump(const Duration(milliseconds: 50));

      await expectLater(
        find.byType(MaterialApp),
        matchesGoldenFile('goldens/onboarding_screen.png'),
      );
    });
  });
}
