import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/settings_provider.dart';
import '../services/auth_service.dart';
import '../services/websocket_service.dart';
import '../utils/constants.dart';

/// Singleton [WebSocketService]; connects on creation and disconnects on dispose.
/// Reads settings (server URL override, language) and the Supabase JWT once at
/// creation so the token is included in the first WebSocket handshake.
///
/// When [AuthService.currentAccessToken] is non-null the server validates it
/// against SUPABASE_JWT_SECRET; in dev mode (no secret configured) the token
/// param is present but ignored by the server.
final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  final settings = ref.read(settingsProvider);
  final baseUrl = settings.serverUrlOverride ?? AppConstants.wsBaseUrl;
  // Watch only language so the service reconnects when the user changes it.
  // Other settings (voice, theme) must not trigger a reconnect.
  final language =
      ref.watch(settingsProvider.select((s) => s.language));

  final service = WebSocketService(
    baseUrl,
    token: AuthService.currentAccessToken,
    language: language,
  );
  service.connect();
  ref.onDispose(() => service.disconnect());
  return service;
});

/// Reactive stream of the current [ConnectionStatus].
final connectionStatusProvider = StreamProvider<ConnectionStatus>((ref) {
  return ref.watch(webSocketServiceProvider).statusStream;
});
