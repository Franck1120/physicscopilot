// Flutter integration tests — require a connected device or emulator.
// Run with: flutter test integration_test/
//
// These tests exercise the full navigation flow of PhysicsCopilotApp by
// mounting the real app widget with controlled state (SharedPreferences,
// equipment provider) rather than calling main() directly, which avoids
// flaky platform-channel interactions during CI.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart';
import 'package:physicscopilot/providers/equipment_provider.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/utils/strings.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  // ---------------------------------------------------------------------------
  // Helpers
  // ---------------------------------------------------------------------------

  const testEquipment = EquipmentProfile(
    id: 'test-printer',
    name: 'Test Printer X1',
    manufacturer: 'TestBrand',
    extruderType: 'direct',
    enclosed: false,
  );

  Future<void> pumpApp(
    WidgetTester tester,
    SharedPreferences prefs, {
    EquipmentProfile? equipment,
  }) async {
    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          if (equipment != null)
            equipmentProvider.overrideWith((ref) {
              final notifier = EquipmentNotifier();
              notifier.select(equipment);
              return notifier;
            }),
        ],
        child: PhysicsCopilotApp(prefs: prefs),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Tests
  // ---------------------------------------------------------------------------

  group('App E2E — splash screen', () {
    testWidgets('renders science icon and app title immediately', (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      await pumpApp(tester, prefs);
      await tester.pump();

      // The splash screen shows the science icon and the app title.
      expect(find.byIcon(Icons.science_rounded), findsWidgets);
      expect(find.text(AppStrings.appName), findsWidgets);

      // Advance past the 2 400 ms splash timer to prevent pending-timer warnings.
      await tester.pump(const Duration(milliseconds: 2500));
      await tester.pump();
    });

    testWidgets('shows loading indicator while transitioning', (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      await pumpApp(tester, prefs);
      await tester.pump(const Duration(milliseconds: 800));

      expect(find.byType(CircularProgressIndicator), findsWidgets);

      await tester.pump(const Duration(milliseconds: 2000));
      await tester.pump();
    });
  });

  group('App E2E — first-launch flow', () {
    testWidgets('navigates to onboarding when not yet completed', (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      await pumpApp(tester, prefs);
      // Jump past the 2 400 ms splash timer.
      await tester.pump(const Duration(milliseconds: 2500));
      await tester.pump();

      // First launch with empty prefs → redirect guard sends user to /onboarding.
      expect(find.text('Punta la camera'), findsOneWidget);
      expect(find.text('Avanti'), findsOneWidget);
    });

    testWidgets('onboarding next button advances to second slide', (tester) async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      await pumpApp(tester, prefs);
      await tester.pump(const Duration(milliseconds: 2500));
      await tester.pump();

      // Tap Avanti to go to the second onboarding slide.
      await tester.tap(find.text('Avanti'));
      await tester.pumpAndSettle();

      expect(find.text("L'AI analizza"), findsOneWidget);
    });
  });

  group('App E2E — home screen', () {
    testWidgets('shows main CTA and bottom nav when setup complete', (tester) async {
      SharedPreferences.setMockInitialValues({
        'onboarding_completed': true,
      });
      final prefs = await SharedPreferences.getInstance();

      await pumpApp(tester, prefs, equipment: testEquipment);
      await tester.pump(const Duration(milliseconds: 2500));
      await tester.pump(const Duration(milliseconds: 200));

      // Home tab: prominent "Nuova sessione" card + bottom nav bar.
      expect(find.text('Nuova sessione'), findsOneWidget);
      expect(find.byType(BottomNavigationBar), findsOneWidget);
      expect(find.text('Home'), findsOneWidget);
      expect(find.text('Cronologia'), findsOneWidget);
    });

    testWidgets('shows active device in equipment section', (tester) async {
      SharedPreferences.setMockInitialValues({
        'onboarding_completed': true,
      });
      final prefs = await SharedPreferences.getInstance();

      await pumpApp(tester, prefs, equipment: testEquipment);
      await tester.pump(const Duration(milliseconds: 2500));
      await tester.pump(const Duration(milliseconds: 200));

      expect(find.text('Test Printer X1'), findsOneWidget);
    });
  });

  group('App E2E — settings navigation', () {
    testWidgets('navigates to settings from profile tab and returns', (tester) async {
      SharedPreferences.setMockInitialValues({
        'onboarding_completed': true,
      });
      final prefs = await SharedPreferences.getInstance();

      await pumpApp(tester, prefs, equipment: testEquipment);
      await tester.pump(const Duration(milliseconds: 2500));
      await tester.pump(const Duration(milliseconds: 200));

      // Tap the Profile tab (4th item, label "Profilo").
      await tester.tap(find.text('Profilo'));
      await tester.pumpAndSettle();

      // Tap "Impostazioni" to navigate to SettingsScreen.
      await tester.tap(find.text('Impostazioni'));
      await tester.pumpAndSettle();

      // SettingsScreen renders the server URL field.
      expect(find.text('URL Server'), findsOneWidget);
      expect(find.text('Guida vocale'), findsOneWidget);

      // Go back to the profile tab.
      await tester.pageBack();
      await tester.pumpAndSettle();

      expect(find.text('Profilo'), findsWidgets);
    });
  });

  group('App E2E — history navigation', () {
    testWidgets('navigates to history tab and back to home', (tester) async {
      SharedPreferences.setMockInitialValues({
        'onboarding_completed': true,
      });
      final prefs = await SharedPreferences.getInstance();

      await pumpApp(tester, prefs, equipment: testEquipment);
      await tester.pump(const Duration(milliseconds: 2500));
      await tester.pump(const Duration(milliseconds: 200));

      // Tap the Cronologia (history) tab — index 2.
      await tester.tap(find.text('Cronologia'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      // History tab is rendered inside the IndexedStack of HomeScreen.
      // The tab label appears in both the nav bar and as a screen heading.
      expect(find.text('Cronologia'), findsWidgets);

      // Return to the Home tab.
      await tester.tap(find.text('Home'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));

      expect(find.text('Nuova sessione'), findsOneWidget);
    });
  });
}
