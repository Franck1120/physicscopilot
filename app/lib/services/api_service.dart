import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/settings_provider.dart';
import '../utils/constants.dart';

/// REST API client for the PhysicsCopilot Go server.
///
/// Uses Dio with conservative timeouts and a single automatic retry
/// (after 1.5 s) to handle transient network hiccups without hammering
/// the server.
class ApiService {
  ApiService({required String baseUrl})
      : _dio = Dio(
          BaseOptions(
            baseUrl: baseUrl,
            connectTimeout: const Duration(seconds: 5),
            receiveTimeout: const Duration(seconds: 10),
          ),
        );

  final Dio _dio;

  /// Returns `true` when the server responds with HTTP 200 to `GET /health`.
  ///
  /// Never throws — all errors are swallowed and expressed as `false`.
  /// Retries once after 1.5 s on any failure before giving up.
  Future<bool> healthCheck() async {
    for (var attempt = 0; attempt < 2; attempt++) {
      if (attempt > 0) {
        await Future.delayed(const Duration(milliseconds: 1500));
      }
      try {
        final response = await _dio.get<void>('/health');
        if (response.statusCode == 200) return true;
      } catch (_) {
        // Swallow DioException, SocketException, etc.
      }
    }
    return false;
  }
}

/// Provides an [ApiService] wired to the user's runtime server-URL override,
/// falling back to the compile-time [AppConstants.apiBaseUrl].
final apiServiceProvider = Provider<ApiService>((ref) {
  final settings = ref.watch(settingsProvider);
  final baseUrl = settings.serverUrlOverride ?? AppConstants.apiBaseUrl;
  return ApiService(baseUrl: baseUrl);
});

/// Polls `GET /health` every 15 seconds and emits the result as a `bool`.
///
/// Consumers can use this to show a live server-status indicator without
/// wiring polling logic into the UI layer.
final serverHealthProvider = StreamProvider<bool>((ref) async* {
  final api = ref.watch(apiServiceProvider);
  while (true) {
    yield await api.healthCheck();
    await Future.delayed(const Duration(seconds: 15));
  }
});
