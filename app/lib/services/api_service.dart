import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/settings_provider.dart';
import '../utils/constants.dart';

// ---------------------------------------------------------------------------
// Remote session model — mirrors the Go server's SessionState JSON response.
// Distinct from SessionRecord, which is persisted locally in SharedPreferences.
// ---------------------------------------------------------------------------

/// Represents an in-memory repair session as returned by the Go server.
class RemoteSession {
  final String sessionId;
  final String deviceBrand;
  final String deviceModel;
  final String problemDetected;
  final DateTime createdAt;
  final DateTime lastActivity;
  final int currentStep;
  final int totalSteps;

  const RemoteSession({
    required this.sessionId,
    required this.deviceBrand,
    required this.deviceModel,
    required this.problemDetected,
    required this.createdAt,
    required this.lastActivity,
    required this.currentStep,
    required this.totalSteps,
  });

  factory RemoteSession.fromJson(Map<String, dynamic> json) {
    final deviceInfo = json['device_info'] as Map<String, dynamic>? ?? {};
    return RemoteSession(
      sessionId: json['session_id'] as String? ?? '',
      deviceBrand: deviceInfo['brand'] as String? ?? '',
      deviceModel: deviceInfo['model'] as String? ?? '',
      problemDetected: json['problem_detected'] as String? ?? '',
      createdAt:
          DateTime.tryParse(json['created_at'] as String? ?? '') ??
          DateTime.now(),
      lastActivity:
          DateTime.tryParse(json['last_activity'] as String? ?? '') ??
          DateTime.now(),
      currentStep: json['current_step'] as int? ?? 0,
      totalSteps: json['total_steps'] as int? ?? 0,
    );
  }

  /// Display name combining brand and model.
  String get deviceName =>
      [deviceBrand, deviceModel].where((s) => s.isNotEmpty).join(' ');
}

// ---------------------------------------------------------------------------
// ApiService — REST client for the PhysicsCopilot Go server.
// ---------------------------------------------------------------------------

/// REST API client for the PhysicsCopilot Go server.
///
/// Uses Dio with conservative timeouts and a single automatic retry
/// (after 1.5 s) to handle transient network hiccups without hammering
/// the server.
///
/// Pass [token] (Supabase JWT) to include an `Authorization: Bearer …`
/// header on every request. When null, no auth header is sent — the
/// server's REST endpoints do not currently require authentication.
class ApiService {
  ApiService({required String baseUrl, this.token})
      : _dio = Dio(
          BaseOptions(
            baseUrl: baseUrl,
            connectTimeout: const Duration(seconds: 5),
            receiveTimeout: const Duration(seconds: 10),
          ),
        );

  final Dio _dio;

  /// Optional JWT for authenticated requests.
  final String? token;

  Options get _opts => token != null
      ? Options(headers: {'Authorization': 'Bearer $token'})
      : Options();

  // ── Health ──────────────────────────────────────────────────────────────

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

  // ── Sessions CRUD ────────────────────────────────────────────────────────

  /// POST /api/sessions — creates a new in-memory repair session.
  ///
  /// Returns the created [RemoteSession] or `null` on any error.
  Future<RemoteSession?> createSession({
    required String deviceBrand,
    required String deviceModel,
  }) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/api/sessions',
        data: {'device_brand': deviceBrand, 'device_model': deviceModel},
        options: _opts,
      );
      if (response.statusCode == 201 && response.data != null) {
        return RemoteSession.fromJson(response.data!);
      }
    } catch (_) {}
    return null;
  }

  /// GET /api/sessions — lists all active in-memory sessions.
  ///
  /// Returns an empty list on any error.
  Future<List<RemoteSession>> listSessions() async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/api/sessions',
        options: _opts,
      );
      if (response.statusCode == 200 && response.data != null) {
        final list = response.data!['sessions'] as List<dynamic>? ?? [];
        return list
            .whereType<Map<String, dynamic>>()
            .map(RemoteSession.fromJson)
            .toList();
      }
    } catch (_) {}
    return [];
  }

  /// GET /api/sessions/:id — fetches a single session by ID.
  ///
  /// Returns `null` when not found or on any error.
  Future<RemoteSession?> getSession(String id) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/api/sessions/$id',
        options: _opts,
      );
      if (response.statusCode == 200 && response.data != null) {
        return RemoteSession.fromJson(response.data!);
      }
    } catch (_) {}
    return null;
  }

  /// DELETE /api/sessions/:id — removes a session from the server.
  ///
  /// Returns `true` on HTTP 204, `false` otherwise.
  Future<bool> deleteSession(String id) async {
    try {
      final response = await _dio.delete<void>(
        '/api/sessions/$id',
        options: _opts,
      );
      return response.statusCode == 204;
    } catch (_) {
      return false;
    }
  }

  // ── Feedback ─────────────────────────────────────────────────────────────

  /// POST /api/feedback — submits user rating for an AI response.
  ///
  /// Body: {"session_id": "...", "step_number": 0, "rating": "up"|"down"}
  /// Returns true on HTTP 201/200, false on any error.
  /// Never throws.
  Future<bool> submitFeedback({
    required String sessionId,
    required int stepNumber,
    required bool liked,
  }) async {
    try {
      final response = await _dio.post<void>(
        '/api/feedback',
        data: {
          'session_id': sessionId,
          'step_number': stepNumber,
          'rating': liked ? 'up' : 'down',
        },
        options: _opts,
      );
      return response.statusCode == 201 || response.statusCode == 200;
    } catch (_) {
      return false;
    }
  }
}

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

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
