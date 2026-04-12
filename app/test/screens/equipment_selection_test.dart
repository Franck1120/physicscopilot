// Widget tests for EquipmentSelectionScreen.
//
// The screen loads printer_profiles.json via rootBundle and uses
// equipmentProvider (Riverpod).  We provide the JSON via a test asset
// bundle and override the provider so no real data is needed.
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/providers/equipment_provider.dart';
import 'package:physicscopilot/screens/equipment_selection_screen.dart';

// ---------------------------------------------------------------------------
// Minimal test asset bundle — returns a tiny JSON for printer_profiles.json
// ---------------------------------------------------------------------------

class _FakeAssetBundle extends CachingAssetBundle {
  @override
  Future<ByteData> load(String key) async {
    if (key == 'assets/data/printer_profiles.json') {
      const json = '''
{
  "profiles": [
    {
      "id": "ender3",
      "name": "Ender 3",
      "manufacturer": "Creality",
      "extruder_type": "bowden",
      "enclosed": false
    },
    {
      "id": "mk4",
      "name": "MK4",
      "manufacturer": "Prusa",
      "extruder_type": "direct",
      "enclosed": false
    }
  ]
}''';
      final bytes = utf8.encode(json);
      return ByteData.view(
          bytes.buffer, bytes.offsetInBytes, bytes.lengthInBytes);
    }
    return super.load(key);
  }
}

void main() {
  group('EquipmentSelectionScreen', () {
    bool completed = false;

    Widget buildTestWidget() {
      completed = false;
      return ProviderScope(
        overrides: [
          equipmentProvider.overrideWith((ref) => EquipmentNotifier()),
        ],
        child: DefaultAssetBundle(
          bundle: _FakeAssetBundle(),
          child: MaterialApp(
            home: EquipmentSelectionScreen(
              onComplete: () => completed = true,
            ),
          ),
        ),
      );
    }

    testWidgets('renders without crash', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pump(); // allow initState to start
      expect(find.byType(EquipmentSelectionScreen), findsOneWidget);
    });

    testWidgets('shows AppBar with title "Seleziona dispositivo"',
        (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pump();
      expect(find.text('Seleziona dispositivo'), findsOneWidget);
    });

    testWidgets('shows loading indicator initially', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      // Before async _loadProfiles completes, CircularProgressIndicator shows.
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows search bar after profiles load', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();
      // Search TextField is visible.
      expect(find.byType(TextField), findsOneWidget);
    });

    testWidgets('shows equipment cards after profiles load', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();
      expect(find.text('Ender 3'), findsOneWidget);
      expect(find.text('MK4'), findsOneWidget);
    });

    testWidgets('shows "Altro dispositivo" custom card', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();
      expect(find.text('Altro dispositivo'), findsOneWidget);
    });

    testWidgets('manufacturer names are shown in cards', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();
      expect(find.text('Creality'), findsOneWidget);
      expect(find.text('Prusa'), findsOneWidget);
    });

    testWidgets('tapping an equipment card calls onComplete', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();
      await tester.tap(find.text('Ender 3'));
      await tester.pump();
      expect(completed, isTrue);
    });

    testWidgets('search field filters equipment list', (tester) async {
      await tester.pumpWidget(buildTestWidget());
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField), 'Ender');
      await tester.pump();
      expect(find.text('Ender 3'), findsOneWidget);
      // MK4 should be filtered out.
      expect(find.text('MK4'), findsNothing);
    });
  });
}
