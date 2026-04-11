import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../services/websocket_service.dart';
import '../utils/constants.dart';

// TODO(auth): replace with Supabase.instance.client.auth.currentSession?.accessToken
// once Supabase Auth is integrated. This token is valid for the default test
// JWT secret ("super-secret-jwt-token-with-at-least-32-characters-long") and
// expires 2030-01-01. Never use this in production.
const _testJwtToken =
    'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9'
    '.eyJlbWFpbCI6InRlc3RAcGh5c2ljc2NvcGlsb3QuZGV2IiwiZXhwIjoxODkzNDU2MDAwLCJpYXQiOjE3NzU5NDY1OTUsInJvbGUiOiJhdXRoZW50aWNhdGVkIiwic3ViIjoidGVzdC11c2VyLWlkLTAwMDAwMDAwMDAwMCJ9'
    '.FBG9ror5pkob-Ng8zfQdqc2epT62K1rzW2PqN7vmrfY';

/// Singleton [WebSocketService]; connects on creation and disconnects on dispose.
final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  final service = WebSocketService(
    AppConstants.wsBaseUrl,
    // getToken is called on every connect/reconnect — swap out _testJwtToken
    // for a real session token once Supabase Auth is wired up.
    getToken: () => _testJwtToken,
  );
  service.connect();
  ref.onDispose(() => service.disconnect());
  return service;
});

/// Reactive stream of the current [ConnectionStatus].
///
/// Resolves to [AsyncValue.data] each time the status changes.
final connectionStatusProvider = StreamProvider<ConnectionStatus>((ref) {
  return ref.watch(webSocketServiceProvider).statusStream;
});
