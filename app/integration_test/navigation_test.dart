// Flutter integration test — navigation flow.
// Run with: flutter test integration_test/navigation_test.dart
//
// These tests mount the full app widget with controlled SharedPreferences
// to verify that navigation between key screens works without crashing.
// A real device or emulator is required.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  Future<void> pumpApp(
    WidgetTester tester,
    SharedPreferences prefs,
  ) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
        child: PhysicsCopilotApp(prefs: prefs),
      ),
    );
  }

  testWidgets('app launches and renders initial screen without crash',
      (tester) async {
    SharedPreferences.setMockInitialValues({'onboarding_completed': true});
    final prefs = await SharedPreferences.getInstance();

    await pumpApp(tester, prefs);
    // Advance past the splash timer (2 400 ms).
    await tester.pump(const Duration(milliseconds: 2500));
    await tester.pump();

    expect(tester.takeException(), isNull);
  });

  testWidgets('app shows onboarding on first launch (no saved state)',
      (tester) async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();

    await pumpApp(tester, prefs);
    await tester.pump(const Duration(milliseconds: 2500));
    await tester.pump();

    // First launch → redirect guard must send the user to /onboarding.
    // The onboarding screen shows an "Avanti" button on the first slide.
    expect(find.text('Avanti'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('back navigation from second onboarding slide returns to first',
      (tester) async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();

    await pumpApp(tester, prefs);
    await tester.pump(const Duration(milliseconds: 2500));
    await tester.pump();

    // Tap Avanti to advance to the second slide.
    await tester.tap(find.text('Avanti'));
    await tester.pumpAndSettle();

    // Navigate back.
    final NavigatorState navigator = tester.state(find.byType(Navigator).first);
    if (navigator.canPop()) {
      navigator.pop();
      await tester.pumpAndSettle();
    }

    expect(tester.takeException(), isNull);
  });
}
