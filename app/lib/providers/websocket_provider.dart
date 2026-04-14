import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/settings_provider.dart';
import '../services/auth_service.dart';
import '../services/websocket_service.dart';
import '../utils/constants.dart';

/// Converts any server URL override to a valid `wss://` base URL.
///
/// Accepts bare host (`physicscopilot.onrender.com`), `https://host`, or
/// `wss://host` — so the user can type any sensible format in Settings.
String _toWssBaseUrl(String raw) {
  if (raw.startsWith('https://')) return 'wss://${raw.substring(8)}';
  if (raw.startsWith('http://')) return 'ws://${raw.substring(7)}';
  if (raw.startsWith('wss://') || raw.startsWith('ws://')) return raw;
  return 'wss://$raw'; // bare host
}

/// Singleton [WebSocketService]; connects on creation and disconnects on dispose.
/// Watches both the server URL override and the language so the service
/// reconnects automatically when either changes.
///
/// When [AuthService.currentAccessToken] is non-null the server validates it
/// against SUPABASE_JWT_SECRET; in dev mode (no secret configured) the token
/// param is present but ignored by the server.
final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  // Watch serverUrlOverride so the service reconnects when the URL changes.
  final overrideUrl =
      ref.watch(settingsProvider.select((s) => s.serverUrlOverride));
  final language =
      ref.watch(settingsProvider.select((s) => s.language));

  final rawBase = overrideUrl ?? AppConstants.wsBaseUrl;
  final baseUrl = _toWssBaseUrl(rawBase);

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
