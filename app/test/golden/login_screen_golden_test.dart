// Golden tests for LoginScreen.
// Run with: flutter test test/golden/login_screen_golden_test.dart --update-goldens
//
// LoginScreen uses Supabase for auth. In golden tests no real Supabase
// instance is needed — we only pump the widget tree and capture the initial
// render (the form fields and buttons) before any user interaction.
//
// Golden files are stored in test/golden/goldens/ and must be committed.
// These snapshots are platform-specific; re-generate on the target CI OS.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart' show sharedPrefsProvider;
import 'package:physicscopilot/screens/login_screen.dart';

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

  group('LoginScreen golden', () {
    late SharedPreferences prefs;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      prefs = await SharedPreferences.getInstance();
    });

    testWidgets('LoginScreen — initial render', (tester) async {
      _fixViewport(tester);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
          child: const MaterialApp(
            debugShowCheckedModeBanner: false,
            home: LoginScreen(),
          ),
        ),
      );
      // Single pump — captures the static form before any async work or
      // Supabase calls that would require a live instance.
      await tester.pump();

      await expectLater(
        find.byType(MaterialApp),
        matchesGoldenFile('goldens/login_screen.png'),
      );
    });
  });
}
