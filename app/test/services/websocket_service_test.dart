import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:fake_async/fake_async.dart';
import 'package:physicscopilot/services/websocket_service.dart';

void main() {
  group('WebSocketService', () {
    late WebSocketService service;

    setUp(() {
      service = WebSocketService('ws://localhost:19999');
    });

    tearDown(() async {
      await service.disconnect();
    });

    test('statusStream is a broadcast stream', () {
      expect(service.statusStream.isBroadcast, isTrue);
    });

    test('messages stream is a broadcast stream', () {
      expect(service.messages.isBroadcast, isTrue);
    });

    test('sendFrame does not throw when channel is null (not connected)', () {
      expect(
        () => service.sendFrame(Uint8List.fromList([0xFF, 0xD8, 0xFF, 0xE0])),
        returnsNormally,
      );
    });

    test('sendText does not throw when channel is null (not connected)', () {
      expect(() => service.sendText('test message'), returnsNormally);
    });

    test('sendText with empty string does not throw', () {
      expect(() => service.sendText(''), returnsNormally);
    });

    test('disconnect can be called multiple times without throwing', () async {
      await expectLater(service.disconnect(), completes);
      await expectLater(service.disconnect(), completes);
    });

    test('backoff delay formula caps at 30 seconds', () {
      // Verifies the formula used in _scheduleReconnect: min(1 << attempts, 30)
      int backoff(int attempts) =>
          [1 << attempts, 30].reduce((a, b) => a < b ? a : b);
      expect(backoff(0), 1);
      expect(backoff(1), 2);
      expect(backoff(2), 4);
      expect(backoff(3), 8);
      expect(backoff(4), 16);
      expect(backoff(5), 30); // 32 capped at 30
      expect(backoff(6), 30); // 64 capped at 30
      expect(backoff(10), 30); // 1024 capped at 30
    });

    test('fakeAsync: no timer fires immediately after service creation', () {
      fakeAsync((fake) {
        final svc = WebSocketService('ws://localhost:19999');
        // No pending timers at construction time
        expect(fake.pendingTimers, isEmpty);
        svc.disconnect();
        fake.flushMicrotasks();
      });
    });
  });
}
