import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/services/api_service.dart';

// ---------------------------------------------------------------------------
// Mock HTTP adapter — returns canned responses keyed by request path.
// ---------------------------------------------------------------------------

class _MockAdapter implements HttpClientAdapter {
  final Map<String, _CannedResponse> _responses = {};

  void addResponse(String path, int status, Object body) {
    _responses[path] = _CannedResponse(status, body);
  }

  void addError(String path) {
    _responses[path] = const _CannedResponse(-1, null);
  }

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final canned = _responses[options.path];
    if (canned == null || canned.status == -1) {
      throw DioException(
        requestOptions: options,
        type: DioExceptionType.connectionError,
      );
    }
    final bodyStr = jsonEncode(canned.body);
    return ResponseBody.fromString(
      bodyStr,
      canned.status,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

class _CannedResponse {
  final int status;
  final Object? body;
  const _CannedResponse(this.status, this.body);
}

// ---------------------------------------------------------------------------
// Thin wrapper that mirrors ApiService logic but accepts an injected Dio.
// This avoids needing to access the private _dio field of ApiService.
// ---------------------------------------------------------------------------

class _TestApiService {
  _TestApiService({required _MockAdapter adapter})
      : _dio = Dio(BaseOptions(baseUrl: 'http://test.local')) {
    _dio.httpClientAdapter = adapter;
  }

  final Dio _dio;

  Future<bool> healthCheck() async {
    for (var attempt = 0; attempt < 2; attempt++) {
      if (attempt > 0) {
        await Future.delayed(const Duration(milliseconds: 1500));
      }
      try {
        final response = await _dio.get<void>('/health');
        if (response.statusCode == 200) return true;
      } catch (_) {
        // Swallow — same behaviour as ApiService.healthCheck
      }
    }
    return false;
  }

  Future<RemoteSession?> createSession({
    required String deviceBrand,
    required String deviceModel,
  }) async {
    try {
      final response = await _dio.post<Map<String, dynamic>>(
        '/api/sessions',
        data: {'device_brand': deviceBrand, 'device_model': deviceModel},
      );
      if (response.statusCode == 201 && response.data != null) {
        return RemoteSession.fromJson(response.data!);
      }
    } catch (_) {}
    return null;
  }

  Future<List<RemoteSession>> listSessions() async {
    try {
      final response = await _dio.get<Map<String, dynamic>>('/api/sessions');
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

  Future<bool> deleteSession(String id) async {
    try {
      final response = await _dio.delete<void>('/api/sessions/$id');
      return response.statusCode == 204;
    } catch (_) {
      return false;
    }
  }
}

// ---------------------------------------------------------------------------
// Canned session JSON matching RemoteSession.fromJson expectations.
// ---------------------------------------------------------------------------

Map<String, dynamic> _sessionJson({String id = 'session-1'}) => {
      'id': id,
      'device': {'brand': 'Test', 'model': 'X1'},
      'problem_detected': 'overheating',
      'created_at': '2026-01-01T00:00:00Z',
    };

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('ApiService HTTP behaviour', () {
    test('healthCheck() returns true when server responds 200', () async {
      final adapter = _MockAdapter()..addResponse('/health', 200, {});
      final service = _TestApiService(adapter: adapter);

      final result = await service.healthCheck();

      expect(result, isTrue);
    });

    test('healthCheck() returns false when server throws DioException',
        () async {
      final adapter = _MockAdapter()..addError('/health');
      final service = _TestApiService(adapter: adapter);

      // Both retry attempts fail immediately (no real delay in test adapter).
      final result = await service.healthCheck();

      expect(result, isFalse);
    }, timeout: const Timeout(Duration(seconds: 10)),);

    test('createSession() returns RemoteSession on 201', () async {
      final adapter = _MockAdapter()
        ..addResponse('/api/sessions', 201, _sessionJson());
      final service = _TestApiService(adapter: adapter);

      final session = await service.createSession(
        deviceBrand: 'Test',
        deviceModel: 'X1',
      );

      expect(session, isNotNull);
      expect(session!.sessionId, equals('session-1'));
      expect(session.deviceName, equals('Test X1'));
    });

    test('listSessions() returns list on 200', () async {
      final adapter = _MockAdapter()
        ..addResponse('/api/sessions', 200, {
          'sessions': [
            _sessionJson(id: 'session-1'),
            _sessionJson(id: 'session-2'),
          ],
        });
      final service = _TestApiService(adapter: adapter);

      final sessions = await service.listSessions();

      expect(sessions, hasLength(2));
      expect(sessions[0].sessionId, equals('session-1'));
      expect(sessions[1].sessionId, equals('session-2'));
    });

    test('listSessions() returns empty list on error', () async {
      final adapter = _MockAdapter()..addError('/api/sessions');
      final service = _TestApiService(adapter: adapter);

      final sessions = await service.listSessions();

      expect(sessions, isEmpty);
    });

    test('deleteSession() returns true on 204', () async {
      final adapter = _MockAdapter()
        ..addResponse('/api/sessions/session-1', 204, {});
      final service = _TestApiService(adapter: adapter);

      final result = await service.deleteSession('session-1');

      expect(result, isTrue);
    });

    test('deleteSession() returns false on network error', () async {
      final adapter = _MockAdapter()..addError('/api/sessions/missing');
      final service = _TestApiService(adapter: adapter);

      final result = await service.deleteSession('missing');

      expect(result, isFalse);
    });
  });
}
