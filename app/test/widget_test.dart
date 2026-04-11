import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/main.dart';

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
    expect(find.byType(PhysicsCopilotApp), findsOneWidget);
  });
}
