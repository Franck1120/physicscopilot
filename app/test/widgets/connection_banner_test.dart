import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/widgets/connection_banner.dart';
import 'package:physicscopilot/services/websocket_service.dart'
    show ConnectionStatus;

void main() {
  Widget wrap(Widget child) =>
      MaterialApp(home: Scaffold(body: child));

  group('ConnectionBanner', () {
    testWidgets('connecting status shows connecting message',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(const ConnectionBanner(status: ConnectionStatus.connecting)),
      );
      await tester.pumpAndSettle();

      expect(find.text('Connessione al server in corso…'), findsOneWidget);
    });

    testWidgets('disconnected status shows unreachable message',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(const ConnectionBanner(status: ConnectionStatus.disconnected)),
      );
      await tester.pumpAndSettle();

      expect(
        find.text('Server non raggiungibile — i dati non vengono inviati'),
        findsOneWidget,
      );
    });

    testWidgets('connecting status has Semantics with liveRegion true',
        (WidgetTester tester) async {
      await tester.pumpWidget(
        wrap(const ConnectionBanner(status: ConnectionStatus.connecting)),
      );
      await tester.pumpAndSettle();

      final semantics = tester.getSemantics(
        find.bySemanticsLabel('Connessione al server in corso…'),
      );
      expect(semantics.flagsCollection.isLiveRegion, isTrue);
    });

    testWidgets('builds without error for all ConnectionStatus values',
        (WidgetTester tester) async {
      for (final status in ConnectionStatus.values) {
        await tester.pumpWidget(
          wrap(ConnectionBanner(status: status)),
        );
        await tester.pumpAndSettle();
        // No exception = pass
        expect(find.byType(ConnectionBanner), findsOneWidget);
      }
    });
  });
}
