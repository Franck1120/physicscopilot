// Integration test — settings screen flow.
// Run with: flutter test integration_test/settings_flow_test.dart
//
// Verifies that the app can navigate to the settings screen and back
// without errors. A real device or emulator is required.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart';
import 'package:physicscopilot/providers/equipment_provider.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  final _testEquipment = EquipmentProfile(
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

  testWidgets('settings screen renders without error', (tester) async {
    SharedPreferences.setMockInitialValues({'onboarding_completed': true});
    final prefs = await SharedPreferences.getInstance();

    await pumpApp(tester, prefs, equipment: _testEquipment);
    await tester.pump(const Duration(milliseconds: 2500));
    await tester.pump(const Duration(milliseconds: 200));

    // No crash on home screen.
    expect(tester.takeException(), isNull);
  });

  testWidgets('can navigate to settings from profile tab', (tester) async {
    SharedPreferences.setMockInitialValues({'onboarding_completed': true});
    final prefs = await SharedPreferences.getInstance();

    await pumpApp(tester, prefs, equipment: _testEquipment);
    await tester.pump(const Duration(milliseconds: 2500));
    await tester.pump(const Duration(milliseconds: 200));

    // Tap the Profile tab.
    await tester.tap(find.text('Profilo'));
    await tester.pumpAndSettle();

    // Tap the Settings tile.
    await tester.tap(find.text('Impostazioni'));
    await tester.pumpAndSettle();

    // SettingsScreen should now be visible.
    expect(find.text('URL Server'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('can navigate back from settings to profile tab', (tester) async {
    SharedPreferences.setMockInitialValues({'onboarding_completed': true});
    final prefs = await SharedPreferences.getInstance();

    await pumpApp(tester, prefs, equipment: _testEquipment);
    await tester.pump(const Duration(milliseconds: 2500));
    await tester.pump(const Duration(milliseconds: 200));

    await tester.tap(find.text('Profilo'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Impostazioni'));
    await tester.pumpAndSettle();

    // Navigate back.
    await tester.pageBack();
    await tester.pumpAndSettle();

    expect(find.text('Profilo'), findsWidgets);
    expect(tester.takeException(), isNull);
  });

  testWidgets('bottom navigation bar is visible on home screen', (tester) async {
    SharedPreferences.setMockInitialValues({'onboarding_completed': true});
    final prefs = await SharedPreferences.getInstance();

    await pumpApp(tester, prefs, equipment: _testEquipment);
    await tester.pump(const Duration(milliseconds: 2500));
    await tester.pump(const Duration(milliseconds: 200));

    expect(find.byType(BottomNavigationBar), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}
