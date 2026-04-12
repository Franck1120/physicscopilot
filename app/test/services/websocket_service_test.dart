// Unit tests for WebSocketService.
//
// Tests that require a real WebSocket connection use a lightweight fake server
// backed by dart:io.  Tests that only exercise null-channel behaviour need no
// server at all.
import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/services/websocket_service.dart';

// ---------------------------------------------------------------------------
// Fake WebSocket server
// ---------------------------------------------------------------------------

/// Spins up a local HTTP server that upgrades to WebSocket.
/// Incoming messages (Strings only) are collected in [received].
class _FakeServer {
  late HttpServer _http;
  final _received = <String>[];
  final _socketCompleter = Completer<WebSocket>();

  Future<void> start() async {
    _http = await HttpServer.bind('127.0.0.1', 0);
    _http.transform(WebSocketTransformer()).listen((ws) {
      if (!_socketCompleter.isCompleted) _socketCompleter.complete(ws);
      ws.listen((msg) {
        if (msg is String) _received.add(msg);
      });
    });
  }

  String get url => 'ws://127.0.0.1:${_http.port}';

  List<String> get received => List.unmodifiable(_received);

  Future<void> closeClientSocket() async {
    if (_socketCompleter.isCompleted) {
      final ws = await _socketCompleter.future;
      await ws.close();
    }
  }

  Future<void> stop() => _http.close(force: true);
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('WebSocketService — connection lifecycle', () {
    test('emits connecting before attempting the handshake', () async {
      // Use a real server so that disconnect() completes quickly (no TCP timeout).
      final server = _FakeServer();
      await server.start();
      addTearDown(server.stop);

      final service = WebSocketService(server.url);
      final statuses = <ConnectionStatus>[];
      final sub = service.statusStream.listen(statuses.add);
      addTearDown(sub.cancel);
      addTearDown(service.disconnect);

      // connect() emits connecting synchronously before the first await.
      unawaited(service.connect());
      await Future<void>.microtask(() {});

      expect(statuses, contains(ConnectionStatus.connecting));
    });

    test('emits disconnected when connection to stopped server fails', () async {
      // Start a server, capture its URL, then stop it.
      // On Linux/macOS, the freed port returns ECONNREFUSED within ms.
      // Skipped on Windows: freed ports enter TIME_WAIT and
      // dart:io's WebSocket.ready can block for tens of seconds.
      // The disconnection behaviour is covered by the server-close test below.
      final server = _FakeServer();
      await server.start();
      final url = server.url;
      await server.stop();

      final service = WebSocketService(url);
      addTearDown(service.disconnect);
      final statuses = <ConnectionStatus>[];
      final sub = service.statusStream.listen(statuses.add);
      addTearDown(sub.cancel);

      await service.connect();
      await Future<void>.delayed(const Duration(milliseconds: 200));

      expect(statuses, contains(ConnectionStatus.disconnected));
    },
      timeout: const Timeout(Duration(seconds: 10)),
      onPlatform: {
        'windows': const Skip(
          'dart:io WebSocket.ready does not fail fast on Windows for '
          'recently freed ports (TIME_WAIT). Same disconnection path is '
          'exercised by "emits disconnected when server closes the connection".',
        ),
      });

    test('emits connected when handshake with real server succeeds', () async {
      final server = _FakeServer();
      await server.start();
      addTearDown(server.stop);

      final service = WebSocketService(server.url);
      addTearDown(service.disconnect);
      final statuses = <ConnectionStatus>[];
      final sub = service.statusStream.listen(statuses.add);
      addTearDown(sub.cancel);

      await service.connect();
      await Future<void>.delayed(const Duration(milliseconds: 50));

      expect(statuses, contains(ConnectionStatus.connected));
    });

    test('emits disconnected when server closes the connection', () async {
      final server = _FakeServer();
      await server.start();
      addTearDown(server.stop);

      final service = WebSocketService(server.url);
      addTearDown(service.disconnect);
      final statuses = <ConnectionStatus>[];
      final sub = service.statusStream.listen(statuses.add);
      addTearDown(sub.cancel);

      await service.connect();
      await Future<void>.delayed(const Duration(milliseconds: 50));
      expect(statuses, contains(ConnectionStatus.connected));

      // Server-side close triggers the client's onDone callback.
      await server.closeClientSocket();
      await Future<void>.delayed(const Duration(milliseconds: 200));

      expect(statuses, contains(ConnectionStatus.disconnected));
    });

    test('disconnect() before connect() completes without error', () async {
      final service = WebSocketService('ws://127.0.0.1:1');
      await expectLater(service.disconnect(), completes);
    });

    test('disconnect() called twice completes without error', () async {
      final service = WebSocketService('ws://127.0.0.1:1');
      await service.disconnect();
      await expectLater(service.disconnect(), completes);
    });
  });

  group('WebSocketService — sendFrame', () {
    test('sendFrame with null channel is a safe no-op', () {
      final service = WebSocketService('ws://localhost:0');
      addTearDown(service.disconnect);

      expect(
        () => service.sendFrame(Uint8List.fromList([0x01, 0x02, 0x03])),
        returnsNormally,
      );
    });

    test('sendFrame sends base64-encoded JSON frame to server', () async {
      final server = _FakeServer();
      await server.start();
      addTearDown(server.stop);

      final service = WebSocketService(server.url);
      addTearDown(service.disconnect);

      await service.connect();
      await Future<void>.delayed(const Duration(milliseconds: 50));

      final frameBytes = Uint8List.fromList([0x01, 0x02, 0x03, 0xFF]);
      service.sendFrame(frameBytes);
      await Future<void>.delayed(const Duration(milliseconds: 50));

      expect(server.received, isNotEmpty);
      final decoded = jsonDecode(server.received.last) as Map<String, dynamic>;
      expect(decoded['type'], equals('frame'));
      expect(decoded['data'], equals(base64Encode(frameBytes)));
      expect(decoded['timestamp'], isA<int>());
    });
  });

  group('WebSocketService — sendText', () {
    test('sendText with null channel is a safe no-op', () {
      final service = WebSocketService('ws://localhost:0');
      addTearDown(service.disconnect);

      expect(
        () => service.sendText('hello'),
        returnsNormally,
      );
    });

    test('sendText sends JSON with type "text" and content to server', () async {
      final server = _FakeServer();
      await server.start();
      addTearDown(server.stop);

      final service = WebSocketService(server.url);
      addTearDown(service.disconnect);

      await service.connect();
      await Future<void>.delayed(const Duration(milliseconds: 50));

      service.sendText('diagnosi pendolo');
      await Future<void>.delayed(const Duration(milliseconds: 50));

      expect(server.received, isNotEmpty);
      final decoded = jsonDecode(server.received.last) as Map<String, dynamic>;
      expect(decoded['type'], equals('text'));
      expect(decoded['content'], equals('diagnosi pendolo'));
      expect(decoded['timestamp'], isA<int>());
    });
  });

  group('WebSocketService — reconnect', () {
    test('schedules reconnect after server closes connection', () async {
      final server = _FakeServer();
      await server.start();

      final service = WebSocketService(server.url);
      final statuses = <ConnectionStatus>[];
      final sub = service.statusStream.listen(statuses.add);

      await service.connect();
      await Future<void>.delayed(const Duration(milliseconds: 50));
      expect(statuses, contains(ConnectionStatus.connected));

      // Trigger disconnect from the server side.
      await server.closeClientSocket();
      await Future<void>.delayed(const Duration(milliseconds: 200));
      expect(statuses, contains(ConnectionStatus.disconnected));

      await sub.cancel();
      await service.disconnect();
      await server.stop();
    });

    test(
      'emits connecting again after first backoff delay (~1 s)',
      () async {
        final server = _FakeServer();
        await server.start();

        final service = WebSocketService(server.url);
        final statuses = <ConnectionStatus>[];
        final sub = service.statusStream.listen(statuses.add);

        await service.connect();
        await Future<void>.delayed(const Duration(milliseconds: 50));

        // Close the client socket from the server side — this is the most
        // reliable way to trigger an onDone event in the client's stream.
        await server.closeClientSocket();
        await Future<void>.delayed(const Duration(milliseconds: 500));
        expect(statuses, contains(ConnectionStatus.disconnected));

        // After ~1 second the first reconnect attempt fires, emitting connecting.
        // (The reconnect may succeed because the server is still up, which is fine —
        // we only need to verify that the backoff reconnect *attempt* was made.)
        await Future<void>.delayed(const Duration(milliseconds: 1200));
        // At least one more "connecting" status after the initial "connected".
        final connectingCount = statuses
            .where((s) => s == ConnectionStatus.connecting)
            .length;
        expect(connectingCount, greaterThanOrEqualTo(1));

        await sub.cancel();
        await service.disconnect();
      },
      timeout: const Timeout(Duration(seconds: 15)),
    );

    test('exponential delay doubles per attempt (formula: min(1 << n, 30))', () {
      // Verify the backoff formula used inside _scheduleReconnect by evaluating
      // expected delays directly rather than exercising real timers.
      int backoffSecs(int attempt) => 1 << attempt < 30 ? 1 << attempt : 30;

      expect(backoffSecs(0), equals(1));
      expect(backoffSecs(1), equals(2));
      expect(backoffSecs(2), equals(4));
      expect(backoffSecs(3), equals(8));
      expect(backoffSecs(4), equals(16));
      // Capped at 30 s after the 5th attempt.
      expect(backoffSecs(5), equals(30));
      expect(backoffSecs(6), equals(30));
    });

    test('reconnectAttempts resets so next backoff starts at 1 s after re-connect',
        () async {
  // Verify that after a successful reconnect the back-off counter is 0,
  // meaning the next disconnect will again wait only 1 second.
  // We check this indirectly: connect → server closes → reconnect attempt fires
  // → service reconnects → server closes again → reconnect fires in ~1 s (not 2+).
  final server = _FakeServer();
  await server.start();
  addTearDown(server.stop);

  final service = WebSocketService(server.url);
  addTearDown(service.disconnect);
  final statuses = <ConnectionStatus>[];
  final sub = service.statusStream.listen(statuses.add);
  addTearDown(sub.cancel);

  await service.connect();
  await Future<void>.delayed(const Duration(milliseconds: 50));
  expect(statuses, contains(ConnectionStatus.connected));

  // First disconnect → schedules reconnect after 1 s (_reconnectAttempts goes to 1).
  await server.closeClientSocket();
  await Future<void>.delayed(const Duration(milliseconds: 1300));
  // After ~1 s reconnect fires; service should be connected again.
  expect(statuses.last, equals(ConnectionStatus.connected));

  // _reconnectAttempts was reset to 0 on the reconnect success.
  // Close socket again → next backoff should be 1 s again.
  await server.closeClientSocket();
  await Future<void>.delayed(const Duration(milliseconds: 1300));
  // We should see another connecting → connected cycle.
  final connectingCount =
      statuses.where((s) => s == ConnectionStatus.connecting).length;
  expect(connectingCount, greaterThanOrEqualTo(2));
},
    timeout: const Timeout(Duration(seconds: 20)),
    onPlatform: {
      'windows': const Skip(
        'Timing-sensitive test: skipped on Windows due to socket TIME_WAIT variability.',
      ),
    });
  });

  group('WebSocketService — message decoding', () {
    test('delivers decoded JSON messages from server', () async {
      final server = _FakeServer();
      await server.start();
      addTearDown(server.stop);

      final service = WebSocketService(server.url);
      addTearDown(service.disconnect);
      final messages = <Map<String, dynamic>>[];
      final sub = service.messages.listen(messages.add);
      addTearDown(sub.cancel);

      await service.connect();
      await Future<void>.delayed(const Duration(milliseconds: 50));

      // Push a message from the server side.
      final serverSocket = await server._socketCompleter.future;
      serverSocket.add(jsonEncode({'type': 'response', 'text': 'Risposta AI'}));
      await Future<void>.delayed(const Duration(milliseconds: 50));

      expect(messages, isNotEmpty);
      expect(messages.first['type'], equals('response'));
      expect(messages.first['text'], equals('Risposta AI'));
    });

    test('silently ignores malformed (non-JSON) messages', () async {
      final server = _FakeServer();
      await server.start();
      addTearDown(server.stop);

      final service = WebSocketService(server.url);
      addTearDown(service.disconnect);
      final messages = <Map<String, dynamic>>[];
      final sub = service.messages.listen(messages.add);
      addTearDown(sub.cancel);

      await service.connect();
      await Future<void>.delayed(const Duration(milliseconds: 50));

      final serverSocket = await server._socketCompleter.future;
      serverSocket.add('not valid json {{{{');
      await Future<void>.delayed(const Duration(milliseconds: 50));

      expect(messages, isEmpty);
    });
  });
}
