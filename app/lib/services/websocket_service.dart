import 'dart:async';
import 'dart:convert';
import 'dart:math';
import 'dart:typed_data';

import 'package:web_socket_channel/web_socket_channel.dart';

/// Connection state exposed to the rest of the application.
enum ConnectionStatus { disconnected, connecting, connected }

/// Manages a persistent WSS connection to the backend.
///
/// Frames are sent as JSON: `{"type":"frame","data":"<base64>","timestamp":<ms>}`.
/// Incoming messages are decoded and broadcast on [messages].
///
/// On disconnection the service reconnects automatically using exponential
/// back-off capped at 60 seconds.
///
/// Authentication: the JWT token is sent as a `?token=` query parameter on
/// every connection attempt, including reconnects (always reads the latest
/// token so refresh tokens are picked up automatically).
class WebSocketService {
  final String _baseUrl;

  /// Returns the current JWT token for the WebSocket connection.
  /// Called on every connect/reconnect to pick up refreshed tokens.
  /// Return null to connect without a token (server will reject with 401).
  final String? Function() _getToken;

  WebSocketChannel? _channel;
  StreamSubscription<dynamic>? _subscription;
  bool _disposed = false;
  int _reconnectAttempts = 0;

  final _statusController =
      StreamController<ConnectionStatus>.broadcast();
  final _messageController =
      StreamController<Map<String, dynamic>>.broadcast();

  WebSocketService(this._baseUrl, {required String? Function() getToken})
      : _getToken = getToken;

  /// Connection status changes (connecting → connected → disconnected → …).
  Stream<ConnectionStatus> get statusStream => _statusController.stream;

  /// Decoded JSON messages received from the server.
  Stream<Map<String, dynamic>> get messages => _messageController.stream;

  /// Opens the WebSocket connection. Calls [_scheduleReconnect] on failure.
  ///
  /// The JWT token is appended as `?token=<jwt>` on every call, so refreshed
  /// tokens are picked up automatically on reconnects.
  Future<void> connect() async {
    if (_disposed) return;
    _emit(ConnectionStatus.connecting);

    try {
      final token = _getToken();
      final wsUri = Uri.parse('$_baseUrl/ws').replace(
        queryParameters: token != null ? {'token': token} : null,
      );
      _channel = WebSocketChannel.connect(wsUri);
      await _channel!.ready;

      _reconnectAttempts = 0;
      _emit(ConnectionStatus.connected);

      _subscription = _channel!.stream.listen(
        _onData,
        onError: (_) => _scheduleReconnect(),
        onDone: _scheduleReconnect,
        cancelOnError: false,
      );
    } catch (_) {
      _scheduleReconnect();
    }
  }

  void _onData(dynamic raw) {
    if (raw is! String) return;
    try {
      final decoded = jsonDecode(raw);
      if (decoded is Map<String, dynamic> && !_messageController.isClosed) {
        _messageController.add(decoded);
      }
    } catch (_) {
      // Ignore malformed messages.
    }
  }

  void _scheduleReconnect() {
    if (_disposed) return;
    _emit(ConnectionStatus.disconnected);
    // Exponential back-off: 1 s, 2 s, 4 s, 8 s … capped at 30 s.
    final delaySecs = min(1 << _reconnectAttempts, 30);
    _reconnectAttempts++;
    Future.delayed(Duration(seconds: delaySecs), connect);
  }

  /// Encodes [frameBytes] as base64 and sends it to the backend.
  void sendFrame(Uint8List frameBytes) {
    if (_channel == null) return;
    try {
      _channel!.sink.add(jsonEncode({
        'type': 'frame',
        'data': base64Encode(frameBytes),
        'timestamp': DateTime.now().millisecondsSinceEpoch,
      }));
    } catch (_) {
      // Connection may have dropped; reconnect will be triggered by onDone.
    }
  }

  /// Sends a voice/text query to the backend.
  void sendText(String text) {
    if (_channel == null) return;
    try {
      _channel!.sink.add(jsonEncode({
        'type': 'voice',
        'text': text,
        'timestamp': DateTime.now().millisecondsSinceEpoch,
      }));
    } catch (_) {
      // Connection may have dropped; reconnect will be triggered by onDone.
    }
  }

  void _emit(ConnectionStatus status) {
    if (!_statusController.isClosed) _statusController.add(status);
  }

  Future<void> disconnect() async {
    _disposed = true;
    await _subscription?.cancel();
    await _channel?.sink.close();
    if (!_statusController.isClosed) await _statusController.close();
    if (!_messageController.isClosed) await _messageController.close();
  }
}
