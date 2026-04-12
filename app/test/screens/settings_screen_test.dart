// Widget tests for SettingsScreen.
//
// The screen depends on settingsProvider (backed by SharedPreferences) and
// voiceServiceProvider.  We override both so no real I/O is needed.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:physicscopilot/main.dart' show sharedPrefsProvider;
import 'package:physicscopilot/providers/settings_provider.dart';
import 'package:physicscopilot/providers/voice_provider.dart';
import 'package:physicscopilot/services/voice_service.dart';
import 'package:physicscopilot/screens/settings_screen.dart';

void main() {
  group('SettingsScreen', () {
    Future<Widget> buildTestWidget({
      Map<String, Object> prefs = const {},
    }) async {
      SharedPreferences.setMockInitialValues(prefs);
      final sharedPrefs = await SharedPreferences.getInstance();

      return ProviderScope(
        overrides: [
          sharedPrefsProvider.overrideWithValue(sharedPrefs),
          settingsProvider.overrideWith(
            (ref) => SettingsNotifier(sharedPrefs),
          ),
          voiceServiceProvider.overrideWithValue(VoiceService()),
        ],
        child: const MaterialApp(
          home: SettingsScreen(),
        ),
      );
    }

    testWidgets('renders without crash', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.byType(SettingsScreen), findsOneWidget);
    });

    testWidgets('shows AppBar title "Impostazioni"', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.text('Impostazioni'), findsOneWidget);
    });

    testWidgets('shows "URL Server" label in connection section', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.text('URL Server'), findsOneWidget);
    });

    testWidgets('shows server URL text field', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.byType(TextField), findsOneWidget);
    });

    testWidgets('shows "Guida vocale" label in features section', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.text('Guida vocale'), findsOneWidget);
    });

    testWidgets('shows voice toggle switch', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.byType(Switch), findsWidgets);
    });

    testWidgets('shows "Tema scuro" label', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.text('Tema scuro'), findsOneWidget);
    });

    testWidgets('shows section header CONNESSIONE', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.text('CONNESSIONE'), findsOneWidget);
    });

    testWidgets('shows section header FUNZIONALITÀ', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.text('FUNZIONALITÀ'), findsOneWidget);
    });

    testWidgets('shows INFORMAZIONI APP section', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.text('INFORMAZIONI APP'), findsOneWidget);
    });

    testWidgets('shows app name PhysicsCopilot in about section', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      // PhysicsCopilot appears in the about info row.
      expect(find.text('PhysicsCopilot'), findsOneWidget);
    });

    testWidgets('shows language dropdown', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      expect(find.byType(DropdownButton<String>), findsOneWidget);
    });

    testWidgets('Salva button is disabled when URL field is not edited',
        (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      // The ElevatedButton labeled 'Salva' should be disabled initially
      // (_urlEdited == false).
      final saveBtn = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Salva'),
      );
      expect(saveBtn.onPressed, isNull);
    });

    testWidgets('Salva button becomes enabled when URL is edited', (tester) async {
      await tester.pumpWidget(await buildTestWidget());
      await tester.pump();
      await tester.enterText(find.byType(TextField), 'wss://example.com');
      await tester.pump();
      final saveBtn = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, 'Salva'),
      );
      expect(saveBtn.onPressed, isNotNull);
    });
  });
}
