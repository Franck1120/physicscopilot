import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../services/websocket_service.dart';
import '../utils/constants.dart';

/// Singleton [WebSocketService]; connects on creation and disconnects on dispose.
final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  final service = WebSocketService(AppConstants.wsBaseUrl);
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
