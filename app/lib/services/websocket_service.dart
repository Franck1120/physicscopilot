import 'dart:async';
import 'dart:convert';
import 'dart:math';
import 'dart:typed_data';

import 'package:flutter/foundation.dart' show debugPrint;
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
/// When [token] is provided it is appended as `?token=<jwt>` on the WebSocket
/// URL — required by the server's JWT auth middleware when
/// `SUPABASE_JWT_SECRET` is configured. In dev mode (no secret set) the
/// token is ignored by the server.
class WebSocketService {
  final String _baseUrl;

  /// Optional JWT for server-side authentication.
  final String? _token;

  /// BCP-47 language code forwarded to the server as ?lang= query param.
  final String _language;

  WebSocketChannel? _channel;
  StreamSubscription<dynamic>? _subscription;
  bool _disposed = false;
  Timer? _reconnectTimer;
  int _reconnectAttempts = 0;

  final _statusController =
      StreamController<ConnectionStatus>.broadcast();
  final _messageController =
      StreamController<Map<String, dynamic>>.broadcast();

  WebSocketService(this._baseUrl, {String? token, String language = 'it'})
      : _token = token,
        _language = language;

  /// Connection status changes (connecting → connected → disconnected → …).
  Stream<ConnectionStatus> get statusStream => _statusController.stream;

  /// Decoded JSON messages received from the server.
  Stream<Map<String, dynamic>> get messages => _messageController.stream;

  /// Opens the WebSocket connection. Calls [_scheduleReconnect] on failure.
  Future<void> connect() async {
    if (_disposed) return;
    _emit(ConnectionStatus.connecting);

    try {
      final token = _token; // local copy for Dart type promotion
      final queryParams = <String, String>{'lang': _language};
      if (token != null) queryParams['token'] = token;
      final wsUri =
          Uri.parse('$_baseUrl/ws').replace(queryParameters: queryParams);
      _channel = WebSocketChannel.connect(wsUri);
      await _channel!.ready;

      // Reset back-off counter so future disconnections restart from the 1-second delay.
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
    } catch (e) {
      debugPrint('[WS] malformed message ignored: $e\nraw=$raw');
    }
  }

  void _scheduleReconnect() {
    if (_disposed) return;
    _emit(ConnectionStatus.disconnected);
    // Exponential back-off: 1 s, 2 s, 4 s, 8 s … capped at 30 s.
    final delaySecs = min(1 << _reconnectAttempts, 30);
    _reconnectAttempts++;
    _reconnectTimer = Timer(Duration(seconds: delaySecs), () {
      if (!_disposed) connect();
    });
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

  /// Sends a text query to the backend.
  void sendText(String text) {
    if (_channel == null) return;
    try {
      _channel!.sink.add(jsonEncode({
        'type': 'text',
        'content': text,
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
    _reconnectTimer?.cancel();
    _disposed = true;
    await _subscription?.cancel();
    await _channel?.sink.close();
    if (!_statusController.isClosed) await _statusController.close();
    if (!_messageController.isClosed) await _messageController.close();
  }
}
