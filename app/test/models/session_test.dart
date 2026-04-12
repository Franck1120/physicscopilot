import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/models/session.dart';

void main() {
  // ---------------------------------------------------------------------------
  // Helpers
  // ---------------------------------------------------------------------------

  final _createdAt = DateTime(2026, 3, 10, 9, 0, 0);

  Session _makeSession({
    String id = 'sess-001',
    String userId = 'user-abc',
    String? deviceId = 'device-xyz',
    SessionStatus status = SessionStatus.active,
    String? problemDetected,
    String? solutionApplied,
    bool? success,
    int? durationSeconds,
    DateTime? createdAt,
  }) =>
      Session(
        id: id,
        userId: userId,
        deviceId: deviceId,
        status: status,
        problemDetected: problemDetected,
        solutionApplied: solutionApplied,
        success: success,
        durationSeconds: durationSeconds,
        createdAt: createdAt ?? DateTime(2026, 3, 10, 9, 0, 0),
      );

  SessionStep _makeStep({
    String id = 'step-001',
    String sessionId = 'sess-001',
    int stepNumber = 1,
    String instruction = 'Check the paper tray',
    bool verified = false,
    DateTime? createdAt,
  }) =>
      SessionStep(
        id: id,
        sessionId: sessionId,
        stepNumber: stepNumber,
        instruction: instruction,
        verified: verified,
        createdAt: createdAt ?? DateTime(2026, 3, 10, 9, 1, 0),
      );

  // ---------------------------------------------------------------------------
  // SessionStatus enum
  // ---------------------------------------------------------------------------

  group('SessionStatus enum', () {
    test('has exactly three values', () {
      expect(SessionStatus.values.length, 3);
    });

    test('contains active, completed, and abandoned', () {
      expect(
        SessionStatus.values,
        containsAll([
          SessionStatus.active,
          SessionStatus.completed,
          SessionStatus.abandoned,
        ]),
      );
    });
  });

  // ---------------------------------------------------------------------------
  // SessionStatusX extension
  // ---------------------------------------------------------------------------

  group('SessionStatusX.isActive', () {
    test('returns true for active', () {
      expect(SessionStatus.active.isActive, isTrue);
    });

    test('returns false for completed', () {
      expect(SessionStatus.completed.isActive, isFalse);
    });

    test('returns false for abandoned', () {
      expect(SessionStatus.abandoned.isActive, isFalse);
    });
  });

  group('SessionStatusX.toJson', () {
    test('active serializes to "active"', () {
      expect(SessionStatus.active.toJson(), 'active');
    });

    test('completed serializes to "completed"', () {
      expect(SessionStatus.completed.toJson(), 'completed');
    });

    test('abandoned serializes to "abandoned"', () {
      expect(SessionStatus.abandoned.toJson(), 'abandoned');
    });
  });

  group('SessionStatusX.fromJson', () {
    test('parses "active"', () {
      expect(SessionStatusX.fromJson('active'), SessionStatus.active);
    });

    test('parses "completed"', () {
      expect(SessionStatusX.fromJson('completed'), SessionStatus.completed);
    });

    test('parses "abandoned"', () {
      expect(SessionStatusX.fromJson('abandoned'), SessionStatus.abandoned);
    });

    test('falls back to active for unknown string', () {
      expect(SessionStatusX.fromJson('unknown_value'), SessionStatus.active);
    });
  });

  // ---------------------------------------------------------------------------
  // Session.fromJson
  // ---------------------------------------------------------------------------

  group('Session.fromJson', () {
    test('parses minimal required fields', () {
      final json = <String, dynamic>{
        'id': 'sess-001',
        'user_id': 'user-abc',
        'device_id': null,
        'status': 'active',
        'problem_detected': null,
        'solution_applied': null,
        'success': null,
        'duration_seconds': null,
        'created_at': '2026-03-10T09:00:00.000',
      };
      final session = Session.fromJson(json);

      expect(session.id, 'sess-001');
      expect(session.userId, 'user-abc');
      expect(session.deviceId, isNull);
      expect(session.status, SessionStatus.active);
      expect(session.problemDetected, isNull);
      expect(session.success, isNull);
      expect(session.durationSeconds, isNull);
    });

    test('parses all optional fields', () {
      final json = <String, dynamic>{
        'id': 'sess-002',
        'user_id': 'user-xyz',
        'device_id': 'device-001',
        'status': 'completed',
        'problem_detected': 'Paper jam',
        'solution_applied': 'Cleared jam',
        'success': true,
        'duration_seconds': 300,
        'created_at': '2026-03-10T09:00:00.000',
      };
      final session = Session.fromJson(json);

      expect(session.deviceId, 'device-001');
      expect(session.status, SessionStatus.completed);
      expect(session.problemDetected, 'Paper jam');
      expect(session.solutionApplied, 'Cleared jam');
      expect(session.success, isTrue);
      expect(session.durationSeconds, 300);
    });

    test('parses createdAt as DateTime', () {
      final json = <String, dynamic>{
        'id': 'sess-003',
        'user_id': 'u',
        'device_id': null,
        'status': 'active',
        'problem_detected': null,
        'solution_applied': null,
        'success': null,
        'duration_seconds': null,
        'created_at': '2026-03-10T09:00:00.000',
      };
      final session = Session.fromJson(json);
      expect(session.createdAt, DateTime.parse('2026-03-10T09:00:00.000'));
    });
  });

  // ---------------------------------------------------------------------------
  // Session.toJson
  // ---------------------------------------------------------------------------

  group('Session.toJson', () {
    test('serializes all fields', () {
      final session = _makeSession(
        durationSeconds: 180,
        problemDetected: 'Jam',
        solutionApplied: 'Fixed',
        success: false,
        createdAt: _createdAt,
      );
      final json = session.toJson();

      expect(json['id'], 'sess-001');
      expect(json['user_id'], 'user-abc');
      expect(json['device_id'], 'device-xyz');
      expect(json['status'], 'active');
      expect(json['problem_detected'], 'Jam');
      expect(json['solution_applied'], 'Fixed');
      expect(json['success'], isFalse);
      expect(json['duration_seconds'], 180);
      expect(json['created_at'], _createdAt.toIso8601String());
    });

    test('serializes null optional fields as null', () {
      final session = _makeSession(deviceId: null);
      final json = session.toJson();
      expect(json['device_id'], isNull);
      expect(json['problem_detected'], isNull);
      expect(json['success'], isNull);
      expect(json['duration_seconds'], isNull);
    });

    test('round-trips through fromJson → toJson', () {
      final session = _makeSession(
        status: SessionStatus.completed,
        durationSeconds: 60,
        success: true,
      );
      final roundTripped = Session.fromJson(session.toJson());
      expect(roundTripped, session);
    });
  });

  // ---------------------------------------------------------------------------
  // Session.copyWith
  // ---------------------------------------------------------------------------

  group('Session.copyWith', () {
    test('returns identical session when no overrides provided', () {
      final session = _makeSession();
      expect(session.copyWith(), session);
    });

    test('overrides status', () {
      final session = _makeSession();
      final copy = session.copyWith(status: SessionStatus.completed);
      expect(copy.status, SessionStatus.completed);
      expect(copy.id, session.id);
    });

    test('overrides deviceId', () {
      final session = _makeSession();
      final copy = session.copyWith(deviceId: 'new-device');
      expect(copy.deviceId, 'new-device');
    });

    test('overrides durationSeconds', () {
      final session = _makeSession();
      final copy = session.copyWith(durationSeconds: 120);
      expect(copy.durationSeconds, 120);
    });

    test('overrides success', () {
      final session = _makeSession();
      final copy = session.copyWith(success: true);
      expect(copy.success, isTrue);
    });

    test('preserves all unchanged fields', () {
      final session = _makeSession(
        problemDetected: 'Jam',
        solutionApplied: 'Cleared',
      );
      final copy = session.copyWith(id: 'new-id');
      expect(copy.problemDetected, 'Jam');
      expect(copy.solutionApplied, 'Cleared');
      expect(copy.userId, session.userId);
      expect(copy.createdAt, session.createdAt);
    });
  });

  // ---------------------------------------------------------------------------
  // Session equality
  // ---------------------------------------------------------------------------

  group('Session equality', () {
    test('two sessions with identical fields are equal', () {
      final a = _makeSession();
      final b = _makeSession();
      expect(a, equals(b));
    });

    test('sessions with different id are not equal', () {
      final a = _makeSession(id: 'id-1');
      final b = _makeSession(id: 'id-2');
      expect(a, isNot(equals(b)));
    });

    test('sessions with different status are not equal', () {
      final a = _makeSession(status: SessionStatus.active);
      final b = _makeSession(status: SessionStatus.completed);
      expect(a, isNot(equals(b)));
    });

    test('equal sessions have equal hash codes', () {
      final a = _makeSession();
      final b = _makeSession();
      expect(a.hashCode, b.hashCode);
    });

    test('session is equal to itself', () {
      final session = _makeSession();
      expect(session, equals(session));
    });
  });

  // ---------------------------------------------------------------------------
  // Session.isActive getter
  // ---------------------------------------------------------------------------

  group('Session.isActive', () {
    test('returns true when status is active', () {
      final session = _makeSession(status: SessionStatus.active);
      expect(session.isActive, isTrue);
    });

    test('returns false when status is completed', () {
      final session = _makeSession(status: SessionStatus.completed);
      expect(session.isActive, isFalse);
    });

    test('returns false when status is abandoned', () {
      final session = _makeSession(status: SessionStatus.abandoned);
      expect(session.isActive, isFalse);
    });
  });

  // ---------------------------------------------------------------------------
  // Session.duration getter
  // ---------------------------------------------------------------------------

  group('Session.duration', () {
    test('returns Duration from durationSeconds', () {
      final session = _makeSession(durationSeconds: 300);
      expect(session.duration, const Duration(seconds: 300));
    });

    test('returns Duration.zero when durationSeconds is null', () {
      final session = _makeSession(durationSeconds: null);
      expect(session.duration, Duration.zero);
    });

    test('returns correct minutes for large values', () {
      final session = _makeSession(durationSeconds: 3600);
      expect(session.duration.inMinutes, 60);
    });
  });

  // ---------------------------------------------------------------------------
  // SessionStep.fromJson
  // ---------------------------------------------------------------------------

  group('SessionStep.fromJson', () {
    test('parses all fields correctly', () {
      final json = <String, dynamic>{
        'id': 'step-001',
        'session_id': 'sess-001',
        'step_number': 1,
        'instruction': 'Remove the paper tray',
        'verified': true,
        'created_at': '2026-03-10T09:01:00.000',
      };
      final step = SessionStep.fromJson(json);

      expect(step.id, 'step-001');
      expect(step.sessionId, 'sess-001');
      expect(step.stepNumber, 1);
      expect(step.instruction, 'Remove the paper tray');
      expect(step.verified, isTrue);
      expect(step.createdAt, DateTime.parse('2026-03-10T09:01:00.000'));
    });

    test('parses verified as false', () {
      final json = <String, dynamic>{
        'id': 'step-002',
        'session_id': 'sess-001',
        'step_number': 2,
        'instruction': 'Press the reset button',
        'verified': false,
        'created_at': '2026-03-10T09:02:00.000',
      };
      final step = SessionStep.fromJson(json);
      expect(step.verified, isFalse);
    });
  });

  // ---------------------------------------------------------------------------
  // SessionStep.toJson
  // ---------------------------------------------------------------------------

  group('SessionStep.toJson', () {
    test('serializes all fields', () {
      final createdAt = DateTime(2026, 3, 10, 9, 1, 0);
      final step = SessionStep(
        id: 'step-001',
        sessionId: 'sess-001',
        stepNumber: 1,
        instruction: 'Remove the paper tray',
        verified: true,
        createdAt: createdAt,
      );
      final json = step.toJson();

      expect(json['id'], 'step-001');
      expect(json['session_id'], 'sess-001');
      expect(json['step_number'], 1);
      expect(json['instruction'], 'Remove the paper tray');
      expect(json['verified'], isTrue);
      expect(json['created_at'], createdAt.toIso8601String());
    });

    test('round-trips through fromJson → toJson', () {
      final step = _makeStep(verified: true);
      final roundTripped = SessionStep.fromJson(step.toJson());
      expect(roundTripped, step);
    });
  });

  // ---------------------------------------------------------------------------
  // SessionStep.copyWith
  // ---------------------------------------------------------------------------

  group('SessionStep.copyWith', () {
    test('returns identical step when no overrides provided', () {
      final step = _makeStep();
      expect(step.copyWith(), step);
    });

    test('overrides instruction', () {
      final step = _makeStep();
      final copy = step.copyWith(instruction: 'New instruction');
      expect(copy.instruction, 'New instruction');
      expect(copy.id, step.id);
    });

    test('overrides verified', () {
      final step = _makeStep(verified: false);
      final copy = step.copyWith(verified: true);
      expect(copy.verified, isTrue);
    });

    test('overrides stepNumber', () {
      final step = _makeStep(stepNumber: 1);
      final copy = step.copyWith(stepNumber: 5);
      expect(copy.stepNumber, 5);
    });

    test('preserves unchanged fields', () {
      final step = _makeStep();
      final copy = step.copyWith(verified: true);
      expect(copy.sessionId, step.sessionId);
      expect(copy.createdAt, step.createdAt);
    });
  });

  // ---------------------------------------------------------------------------
  // SessionStep equality
  // ---------------------------------------------------------------------------

  group('SessionStep equality', () {
    test('two steps with identical fields are equal', () {
      final a = _makeStep();
      final b = _makeStep();
      expect(a, equals(b));
    });

    test('steps with different id are not equal', () {
      final a = _makeStep(id: 'step-001');
      final b = _makeStep(id: 'step-002');
      expect(a, isNot(equals(b)));
    });

    test('steps with different instruction are not equal', () {
      final a = _makeStep(instruction: 'Step A');
      final b = _makeStep(instruction: 'Step B');
      expect(a, isNot(equals(b)));
    });

    test('steps with different verified are not equal', () {
      final a = _makeStep(verified: false);
      final b = _makeStep(verified: true);
      expect(a, isNot(equals(b)));
    });

    test('equal steps have equal hash codes', () {
      final a = _makeStep();
      final b = _makeStep();
      expect(a.hashCode, b.hashCode);
    });

    test('step is equal to itself', () {
      final step = _makeStep();
      expect(step, equals(step));
    });
  });

  // ---------------------------------------------------------------------------
  // SessionStep.toString
  // ---------------------------------------------------------------------------

  group('SessionStep.toString', () {
    test('includes all field values', () {
      final step = _makeStep(
        id: 'step-001',
        sessionId: 'sess-001',
        stepNumber: 1,
        instruction: 'Check tray',
        verified: false,
      );
      final str = step.toString();

      expect(str, contains('step-001'));
      expect(str, contains('sess-001'));
      expect(str, contains('1'));
      expect(str, contains('Check tray'));
      expect(str, contains('false'));
    });

    test('starts with SessionStep(', () {
      expect(_makeStep().toString(), startsWith('SessionStep('));
    });
  });

  // ---------------------------------------------------------------------------
  // Session.toString
  // ---------------------------------------------------------------------------

  group('Session.toString', () {
    test('includes id, userId, and status', () {
      final session = _makeSession(
        id: 'sess-001',
        userId: 'user-abc',
        status: SessionStatus.active,
      );
      final str = session.toString();

      expect(str, contains('sess-001'));
      expect(str, contains('user-abc'));
      expect(str, contains('SessionStatus.active'));
    });

    test('starts with Session(', () {
      expect(_makeSession().toString(), startsWith('Session('));
    });
  });
}
