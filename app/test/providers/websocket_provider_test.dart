import 'dart:async';
import 'dart:typed_data';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/providers/websocket_provider.dart';
import 'package:physicscopilot/services/websocket_service.dart';

void main() {
  group('WebSocketService', () {
    test('statusStream is a broadcast stream — multiple listeners allowed', () {
      final service = WebSocketService('ws://localhost:0');
      addTearDown(service.disconnect);

      // Subscribing twice must not throw StateError.
      final sub1 = service.statusStream.listen((_) {});
      final sub2 = service.statusStream.listen((_) {});

      expect(sub1, isNotNull);
      expect(sub2, isNotNull);

      sub1.cancel();
      sub2.cancel();
    });

    test('messages stream is a broadcast stream — multiple listeners allowed', () {
      final service = WebSocketService('ws://localhost:0');
      addTearDown(service.disconnect);

      final sub1 = service.messages.listen((_) {});
      final sub2 = service.messages.listen((_) {});

      expect(sub1, isNotNull);
      expect(sub2, isNotNull);

      sub1.cancel();
      sub2.cancel();
    });

    test('disconnect() before connect() completes without error', () async {
      final service = WebSocketService('ws://localhost:0');
      await expectLater(service.disconnect(), completes);
    });

    test('disconnect() called twice completes without error', () async {
      final service = WebSocketService('ws://localhost:0');
      await service.disconnect();
      await expectLater(service.disconnect(), completes);
    });

    test('sendFrame() when not connected is a safe no-op', () {
      final service = WebSocketService('ws://localhost:0');
      addTearDown(service.disconnect);

      expect(
        () => service.sendFrame(Uint8List.fromList([0x01, 0x02, 0x03])),
        returnsNormally,
      );
    });

    test('sendText() when not connected is a safe no-op', () {
      final service = WebSocketService('ws://localhost:0');
      addTearDown(service.disconnect);

      expect(
        () => service.sendText('hello world'),
        returnsNormally,
      );
    });
  });

  group('connectionStatusProvider', () {
    // Override connectionStatusProvider with a controlled stream to isolate
    // provider behaviour from real network calls.

    test('initial state is AsyncLoading before the stream emits', () {
      final controller = StreamController<ConnectionStatus>.broadcast();
      addTearDown(controller.close);

      final container = ProviderContainer(overrides: [
        connectionStatusProvider.overrideWith((ref) => controller.stream),
      ]);
      addTearDown(container.dispose);

      // No value emitted yet → provider must be in loading state.
      expect(container.read(connectionStatusProvider), isA<AsyncLoading>());
    });

    test('emits AsyncData(connected) when stream emits connected', () async {
      final controller = StreamController<ConnectionStatus>.broadcast();

      final container = ProviderContainer(overrides: [
        connectionStatusProvider.overrideWith((ref) => controller.stream),
      ]);

      // Subscribe so the provider activates before we emit.
      final statuses = <ConnectionStatus>[];
      container.listen(connectionStatusProvider, (_, next) {
        if (next is AsyncData<ConnectionStatus>) statuses.add(next.value);
      });

      controller.add(ConnectionStatus.connected);
      await Future<void>.delayed(Duration.zero);

      expect(statuses, contains(ConnectionStatus.connected));

      await controller.close();
      container.dispose();
    });

    test('emits AsyncData(disconnected) when stream emits disconnected', () async {
      final controller = StreamController<ConnectionStatus>.broadcast();

      final container = ProviderContainer(overrides: [
        connectionStatusProvider.overrideWith((ref) => controller.stream),
      ]);

      final statuses = <ConnectionStatus>[];
      container.listen(connectionStatusProvider, (_, next) {
        if (next is AsyncData<ConnectionStatus>) statuses.add(next.value);
      });

      controller.add(ConnectionStatus.disconnected);
      await Future<void>.delayed(Duration.zero);

      expect(statuses, contains(ConnectionStatus.disconnected));

      await controller.close();
      container.dispose();
    });

    test('reflects status transitions: connecting → connected', () async {
      final controller = StreamController<ConnectionStatus>.broadcast();

      final container = ProviderContainer(overrides: [
        connectionStatusProvider.overrideWith((ref) => controller.stream),
      ]);

      final statuses = <ConnectionStatus>[];
      container.listen(connectionStatusProvider, (_, next) {
        if (next is AsyncData<ConnectionStatus>) statuses.add(next.value);
      });

      controller.add(ConnectionStatus.connecting);
      await Future<void>.delayed(Duration.zero);

      controller.add(ConnectionStatus.connected);
      await Future<void>.delayed(Duration.zero);

      expect(statuses, equals([ConnectionStatus.connecting, ConnectionStatus.connected]));

      await controller.close();
      container.dispose();
    });

    test('reflects disconnect after connect: connected → disconnected', () async {
      final controller = StreamController<ConnectionStatus>.broadcast();

      final container = ProviderContainer(overrides: [
        connectionStatusProvider.overrideWith((ref) => controller.stream),
      ]);

      final statuses = <ConnectionStatus>[];
      container.listen(connectionStatusProvider, (_, next) {
        if (next is AsyncData<ConnectionStatus>) statuses.add(next.value);
      });

      controller.add(ConnectionStatus.connected);
      await Future<void>.delayed(Duration.zero);

      controller.add(ConnectionStatus.disconnected);
      await Future<void>.delayed(Duration.zero);

      expect(statuses, equals([ConnectionStatus.connected, ConnectionStatus.disconnected]));

      await controller.close();
      container.dispose();
    });
  });
}
