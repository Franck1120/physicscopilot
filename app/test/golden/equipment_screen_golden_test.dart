// Golden tests for EquipmentSelectionScreen.
// Run with: flutter test test/golden/equipment_screen_golden_test.dart --update-goldens
//
// EquipmentSelectionScreen loads equipment profiles from a JSON asset.
// In golden tests the asset bundle resolves normally; the screen will show
// the search UI in its loading or empty-results state.
//
// Golden files are stored in test/golden/goldens/ and must be committed.
// These snapshots are platform-specific; re-generate on the target CI OS.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/screens/equipment_selection_screen.dart';

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

  group('EquipmentSelectionScreen golden', () {
    late SharedPreferences prefs;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      prefs = await SharedPreferences.getInstance();
    });

    testWidgets('EquipmentSelectionScreen — initial loading state', (tester) async {
      _fixViewport(tester);

      await tester.pumpWidget(
        ProviderScope(
          overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
          child: const MaterialApp(
            debugShowCheckedModeBanner: false,
            home: EquipmentSelectionScreen(onComplete: _noop),
          ),
        ),
      );
      // Single pump — captures the initial loading/skeleton before the async
      // JSON asset load completes.
      await tester.pump();

      await expectLater(
        find.byType(MaterialApp),
        matchesGoldenFile('goldens/equipment_screen.png'),
      );
    });
  });
}
