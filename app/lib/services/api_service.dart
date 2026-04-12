// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/settings_provider.dart';
import '../utils/constants.dart';

// ---------------------------------------------------------------------------
// RemoteSession model — maps GET /api/sessions list response
// ---------------------------------------------------------------------------

/// A session summary returned by `GET /api/sessions`.
///
/// Field names match the server's `sessionResponse` JSON DTO.
class RemoteSession {
  const RemoteSession({
    required this.sessionId,
    required this.deviceName,
    required this.problemDetected,
    required this.createdAt,
  });

  /// Parses one element from the `sessions` array in the list response.
  factory RemoteSession.fromJson(Map<String, dynamic> json) {
    final device = json['device'] as Map<String, dynamic>? ?? {};
    final brand = device['brand'] as String? ?? '';
    final model = device['model'] as String? ?? '';
    final name = [brand, model].where((s) => s.isNotEmpty).join(' ');
    return RemoteSession(
      sessionId: json['id'] as String,
      deviceName: name,
      problemDetected: json['problem_detected'] as String? ?? '',
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  /// Server-assigned session UUID.
  final String sessionId;

  /// Human-readable device name: `"$brand $model"`.
  final String deviceName;

  /// Problem description extracted by the AI, or empty string.
  final String problemDetected;

  /// When the session was created on the server.
  final DateTime createdAt;
}

// ---------------------------------------------------------------------------
// ServerHealth model
// ---------------------------------------------------------------------------

/// Snapshot of the Go server's health state returned by `GET /health`.
///
/// Use [ServerHealth.offline] when the server cannot be reached, and
/// [ServerHealth.fromJson] to parse a successful HTTP 200 response.
class ServerHealth {
  const ServerHealth({
    required this.isOnline,
    required this.version,
    required this.uptimeSeconds,
    required this.activeConnections,
  });

  /// Server is unreachable or returned an error.
  factory ServerHealth.offline() => const ServerHealth(
        isOnline: false,
        version: '',
        uptimeSeconds: 0,
        activeConnections: 0,
      );

  /// Parses the JSON body from a successful `GET /health` response.
  ///
  /// Expects the shape:
  /// ```json
  /// { "status": "ok", "version": "1.0.0",
  ///   "uptime_seconds": 12345, "active_connections": 3 }
  /// ```
  factory ServerHealth.fromJson(Map<String, dynamic> json) => ServerHealth(
        isOnline: true,
        version: json['version'] as String? ?? '',
        uptimeSeconds: json['uptime_seconds'] as int? ?? 0,
        activeConnections: json['active_connections'] as int? ?? 0,
      );

  final bool isOnline;
  final String version;
  final int uptimeSeconds;
  final int activeConnections;

  @override
  String toString() =>
      'ServerHealth(isOnline: $isOnline, version: $version, '
      'uptimeSeconds: $uptimeSeconds, activeConnections: $activeConnections)';
}

// ---------------------------------------------------------------------------
// ApiService
// ---------------------------------------------------------------------------

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

  /// Fetches the list of sessions from `GET /api/sessions`.
  ///
  /// Returns an empty list on any network or parse error (fire-and-forget
  /// pattern — local storage is the source of truth).
  Future<List<RemoteSession>> listSessions() async {
    try {
      final response =
          await _dio.get<Map<String, dynamic>>('/api/sessions');
      if (response.statusCode == 200 && response.data != null) {
        final raw = response.data!['sessions'];
        if (raw is List) {
          return raw
              .whereType<Map<String, dynamic>>()
              .map(RemoteSession.fromJson)
              .toList();
        }
      }
    } catch (_) {}
    return [];
  }

  /// Posts feedback for a session step to `POST /api/feedback`.
  ///
  /// Converts [liked] to `"positive"` or `"negative"` as required by the
  /// server. Fire-and-forget — errors are silently swallowed.
  Future<void> submitFeedback({
    required String sessionId,
    required int stepNumber,
    required bool liked,
  }) async {
    try {
      await _dio.post<void>(
        '/api/feedback',
        data: {
          'session_id': sessionId,
          'step_number': stepNumber,
          'rating': liked ? 'positive' : 'negative',
        },
      );
    } catch (_) {}
  }

  /// Posts a new session record to `POST /api/sessions`.
  ///
  /// Fire-and-forget: errors are silently swallowed because local storage
  /// is the source of truth — server sync is best-effort only.
  Future<void> createSession({
    required String deviceBrand,
    required String deviceModel,
  }) async {
    try {
      await _dio.post<void>(
        '/api/sessions',
        data: {'device_brand': deviceBrand, 'device_model': deviceModel},
      );
    } catch (_) {}
  }

  /// Returns a [ServerHealth] snapshot from `GET /health`.
  ///
  /// Never throws — all errors are expressed as [ServerHealth.offline].
  /// Retries once after 1.5 s on any failure before giving up.
  Future<ServerHealth> healthCheck() async {
    for (var attempt = 0; attempt < 2; attempt++) {
      if (attempt > 0) {
        await Future.delayed(const Duration(milliseconds: 1500));
      }
      try {
        final response = await _dio.get<Map<String, dynamic>>('/health');
        if (response.statusCode == 200 && response.data != null) {
          return ServerHealth.fromJson(response.data!);
        }
      } catch (_) {
        // Swallow DioException, SocketException, etc.
      }
    }
    return ServerHealth.offline();
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

/// Polls `GET /health` every 15 seconds and emits a [ServerHealth] snapshot.
///
/// Consumers that need the full health payload (version, uptime, connections)
/// watch this provider. For a simple online/offline bool, prefer
/// [serverOnlineProvider].
final serverHealthProvider = StreamProvider<ServerHealth>((ref) async* {
  final api = ref.watch(apiServiceProvider);
  while (true) {
    yield await api.healthCheck();
    await Future.delayed(const Duration(seconds: 15));
  }
});

/// Derived `bool` provider — `true` when the server is reachable.
///
/// Convenience wrapper around [serverHealthProvider] for UI widgets that
/// only need to know whether the server is online.
final serverOnlineProvider = Provider<bool>((ref) {
  return ref.watch(serverHealthProvider).valueOrNull?.isOnline ?? false;
});
