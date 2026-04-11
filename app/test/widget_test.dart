import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/main.dart';

void main() {
  testWidgets('App mounts without crashing', (WidgetTester tester) async {
    await tester.pumpWidget(
      const ProviderScope(child: PhysicsCopilotApp()),
    );
    // Just verify the app widget tree is created.
    expect(find.byType(PhysicsCopilotApp), findsOneWidget);
  });
}
