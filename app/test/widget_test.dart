import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/screens/home_screen.dart';
import 'package:physicscopilot/screens/onboarding_screen.dart';
import 'package:physicscopilot/screens/settings_screen.dart';
import 'package:physicscopilot/services/api_service.dart';

void main() {
  testWidgets('App mounts without crashing', (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          // Prevent healthCheck from leaving a pending 1.5 s retry timer.
          serverHealthProvider.overrideWith((_) => Stream.value(false)),
        ],
        child: PhysicsCopilotApp(prefs: prefs),
      ),
    );
    // Splash screen should be visible immediately after mount
    expect(find.byType(PhysicsCopilotApp), findsOneWidget);

    // Advance past the 2400ms splash timer so it fires and the test
    // does not fail with "pending timers" reported by the test framework.
    await tester.pump(const Duration(milliseconds: 2500));
    await tester.pump();
  });

  testWidgets('OnboardingScreen shows first slide and navigates',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
        child: MaterialApp(
          home: OnboardingScreen(onComplete: () {}),
        ),
      ),
    );
    await tester.pumpAndSettle();

    // Verifica titolo prima slide
    expect(find.text('Punta la camera'), findsOneWidget);
    // Verifica pulsante Avanti
    expect(find.text('Avanti'), findsOneWidget);

    // Tap Avanti e verifica seconda slide
    await tester.tap(find.text('Avanti'));
    await tester.pumpAndSettle();
    expect(find.text("L'AI analizza"), findsOneWidget);
  });

  testWidgets('SettingsScreen renders URL field and voice toggle',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
        child: const MaterialApp(
          home: SettingsScreen(),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('URL Server'), findsOneWidget);
    expect(find.text('Guida vocale'), findsOneWidget);
    expect(find.byType(Switch), findsWidgets);
  });

  testWidgets('HomeScreen mounts with nav bar and AppBar title',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          // Prevent healthCheck from leaving a pending 1.5 s retry timer.
          serverHealthProvider.overrideWith((_) => Stream.value(false)),
        ],
        child: MaterialApp(
          home: HomeScreen(
            onChangeEquipment: () {},
            onStartCamera: () {},
          ),
        ),
      ),
    );
    // Drain the 400ms delayed-loading timer inside HomeScreen.
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));

    expect(find.byType(BottomNavigationBar), findsOneWidget);
    expect(find.text('Home'), findsOneWidget);
    expect(find.text('Cronologia'), findsOneWidget);
  });
}
