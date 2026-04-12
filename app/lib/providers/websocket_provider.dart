import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/settings_provider.dart';
import '../services/websocket_service.dart';
import '../utils/constants.dart';

/// Singleton [WebSocketService]; connects on creation and disconnects on dispose.
/// Reads settings once at creation (server URL override and language).
final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  final settings = ref.read(settingsProvider);
  final baseUrl = settings.serverUrlOverride ?? AppConstants.wsBaseUrl;
  final service = WebSocketService(baseUrl, language: settings.language);
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
