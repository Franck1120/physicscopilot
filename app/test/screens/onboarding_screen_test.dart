// Widget tests for OnboardingScreen.
//
// The screen depends on:
//   - sharedPrefsProvider (to persist the completion flag)
//   - onboardingCompletedProvider (StateProvider backed by sharedPrefsProvider)
//
// Permission requests (_requestPermissions) are exercised only when the user
// taps "Inizia" / "Salta". Those interactions are verified at the UI level
// without asserting on permission dialog state (platform-specific).
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/screens/onboarding_screen.dart';

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

/// Builds an [OnboardingScreen] wrapped in a [ProviderScope] and [MaterialApp].
///
/// [onComplete] is called when the screen finishes (skip or last-page button).
Future<Widget> buildOnboardingScreen({
  VoidCallback? onComplete,
}) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();

  return ProviderScope(
    overrides: [
      sharedPrefsProvider.overrideWithValue(prefs),
    ],
    child: MaterialApp(
      home: OnboardingScreen(
        onComplete: onComplete ?? () {},
      ),
    ),
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('OnboardingScreen', () {
    testWidgets('renders without crash', (tester) async {
      await tester.pumpWidget(await buildOnboardingScreen());
      await tester.pump();
      expect(find.byType(OnboardingScreen), findsOneWidget);
    });

    testWidgets('shows first slide title "Punta la camera"', (tester) async {
      await tester.pumpWidget(await buildOnboardingScreen());
      await tester.pump();
      expect(find.text('Punta la camera'), findsOneWidget);
    });

    testWidgets('shows "Avanti" button on first slide', (tester) async {
      await tester.pumpWidget(await buildOnboardingScreen());
      await tester.pump();
      expect(find.text('Avanti'), findsOneWidget);
    });

    testWidgets('shows "Salta" button on first slide', (tester) async {
      await tester.pumpWidget(await buildOnboardingScreen());
      await tester.pump();
      expect(find.text('Salta'), findsOneWidget);
    });

    testWidgets('tapping Avanti navigates to second slide', (tester) async {
      await tester.pumpWidget(await buildOnboardingScreen());
      await tester.pump();

      await tester.tap(find.text('Avanti'));
      // Allow the PageController animation to settle.
      await tester.pumpAndSettle();

      expect(find.text('Segui la voce'), findsOneWidget);
    });

    testWidgets('navigating through all slides shows "Inizia" on last slide',
        (tester) async {
      await tester.pumpWidget(await buildOnboardingScreen());
      await tester.pump();

      // Slide 1 → 2
      await tester.tap(find.text('Avanti'));
      await tester.pumpAndSettle();

      // Slide 2 → 3
      await tester.tap(find.text('Avanti'));
      await tester.pumpAndSettle();

      expect(find.text('Inizia'), findsOneWidget);
    });

    testWidgets('"Salta" button is hidden on the last slide', (tester) async {
      await tester.pumpWidget(await buildOnboardingScreen());
      await tester.pump();

      // Navigate to last slide.
      await tester.tap(find.text('Avanti'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Avanti'));
      await tester.pumpAndSettle();

      expect(find.text('Salta'), findsNothing);
    });

    testWidgets('page indicators render for each slide', (tester) async {
      await tester.pumpWidget(await buildOnboardingScreen());
      await tester.pump();

      // The _PageIndicator generates 3 AnimatedContainers (one per slide).
      expect(find.byType(AnimatedContainer), findsNWidgets(3));
    });

    testWidgets('onComplete callback invoked when "Salta" is tapped',
        (tester) async {
      var called = false;
      await tester.pumpWidget(
        await buildOnboardingScreen(onComplete: () => called = true),
      );
      await tester.pump();

      await tester.tap(find.text('Salta'));
      // Let async _handleComplete work (prefs.setBool is awaited).
      await tester.pump(const Duration(milliseconds: 100));
      await tester.pump();

      expect(called, isTrue);
    });
  });
}
