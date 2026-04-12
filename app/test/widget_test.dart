import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';

void main() {
  testWidgets('App mounts without crashing', (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
        child: PhysicsCopilotApp(prefs: prefs),
      ),
    );
    // Splash screen should be visible immediately after mount
    expect(find.byType(PhysicsCopilotApp), findsOneWidget);

    // Advance past the 2200ms splash timer so it fires and the test
    // does not fail with "pending timers" reported by the test framework.
    await tester.pump(const Duration(milliseconds: 2300));
    await tester.pump();
  });
}
