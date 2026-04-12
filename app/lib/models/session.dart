// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

/// Status of a session: active, completed, or abandoned.
enum SessionStatus {
  active,
  completed,
  abandoned,
}

extension SessionStatusX on SessionStatus {
  /// Returns true if this status is active.
  bool get isActive => this == SessionStatus.active;

  /// Converts enum value to JSON string.
  String toJson() => name;

  /// Parses JSON string to enum value.
  static SessionStatus fromJson(String json) =>
      SessionStatus.values.firstWhere(
        (e) => e.name == json,
        orElse: () => SessionStatus.active,
      );
}

/// Represents a troubleshooting session for a device.
///
/// Maps directly to the `sessions` table in Supabase.
class Session {
  final String id;
  final String userId;
  final String? deviceId;
  final SessionStatus status;
  final String? problemDetected;
  final String? solutionApplied;
  final bool? success;
  final int? durationSeconds;
  final DateTime createdAt;

  const Session({
    required this.id,
    required this.userId,
    this.deviceId,
    required this.status,
    this.problemDetected,
    this.solutionApplied,
    this.success,
    this.durationSeconds,
    required this.createdAt,
  });

  // ---------------------------------------------------------------------------
  // Serialization
  // ---------------------------------------------------------------------------

  factory Session.fromJson(Map<String, dynamic> json) => Session(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        deviceId: json['device_id'] as String?,
        status: SessionStatusX.fromJson(json['status'] as String),
        problemDetected: json['problem_detected'] as String?,
        solutionApplied: json['solution_applied'] as String?,
        success: json['success'] as bool?,
        durationSeconds: json['duration_seconds'] as int?,
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'user_id': userId,
        'device_id': deviceId,
        'status': status.toJson(),
        'problem_detected': problemDetected,
        'solution_applied': solutionApplied,
        'success': success,
        'duration_seconds': durationSeconds,
        'created_at': createdAt.toIso8601String(),
      };

  // ---------------------------------------------------------------------------
  // Convenience
  // ---------------------------------------------------------------------------

  /// Returns true if this session is currently active.
  bool get isActive => status.isActive;

  /// Returns the session duration as a [Duration] object.
  Duration get duration => Duration(seconds: durationSeconds ?? 0);

  // ---------------------------------------------------------------------------
  // copyWith
  // ---------------------------------------------------------------------------

  Session copyWith({
    String? id,
    String? userId,
    String? deviceId,
    SessionStatus? status,
    String? problemDetected,
    String? solutionApplied,
    bool? success,
    int? durationSeconds,
    DateTime? createdAt,
  }) =>
      Session(
        id: id ?? this.id,
        userId: userId ?? this.userId,
        deviceId: deviceId ?? this.deviceId,
        status: status ?? this.status,
        problemDetected: problemDetected ?? this.problemDetected,
        solutionApplied: solutionApplied ?? this.solutionApplied,
        success: success ?? this.success,
        durationSeconds: durationSeconds ?? this.durationSeconds,
        createdAt: createdAt ?? this.createdAt,
      );

  // ---------------------------------------------------------------------------
  // Equality
  // ---------------------------------------------------------------------------

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is Session &&
          runtimeType == other.runtimeType &&
          id == other.id &&
          userId == other.userId &&
          deviceId == other.deviceId &&
          status == other.status &&
          problemDetected == other.problemDetected &&
          solutionApplied == other.solutionApplied &&
          success == other.success &&
          durationSeconds == other.durationSeconds &&
          createdAt == other.createdAt;

  @override
  int get hashCode => Object.hash(
        id,
        userId,
        deviceId,
        status,
        problemDetected,
        solutionApplied,
        success,
        durationSeconds,
        createdAt,
      );

  @override
  String toString() =>
      'Session(id: $id, userId: $userId, deviceId: $deviceId, '
      'status: $status, problemDetected: $problemDetected, '
      'solutionApplied: $solutionApplied, success: $success, '
      'durationSeconds: $durationSeconds, createdAt: $createdAt)';
}

/// Represents a single step in a troubleshooting session.
///
/// Maps directly to the `session_steps` table in Supabase.
class SessionStep {
  final String id;
  final String sessionId;
  final int stepNumber;
  final String instruction;
  final bool verified;
  final DateTime createdAt;

  const SessionStep({
    required this.id,
    required this.sessionId,
    required this.stepNumber,
    required this.instruction,
    required this.verified,
    required this.createdAt,
  });

  // ---------------------------------------------------------------------------
  // Serialization
  // ---------------------------------------------------------------------------

  factory SessionStep.fromJson(Map<String, dynamic> json) => SessionStep(
        id: json['id'] as String,
        sessionId: json['session_id'] as String,
        stepNumber: json['step_number'] as int,
        instruction: json['instruction'] as String,
        verified: json['verified'] as bool,
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'session_id': sessionId,
        'step_number': stepNumber,
        'instruction': instruction,
        'verified': verified,
        'created_at': createdAt.toIso8601String(),
      };

  // ---------------------------------------------------------------------------
  // copyWith
  // ---------------------------------------------------------------------------

  SessionStep copyWith({
    String? id,
    String? sessionId,
    int? stepNumber,
    String? instruction,
    bool? verified,
    DateTime? createdAt,
  }) =>
      SessionStep(
        id: id ?? this.id,
        sessionId: sessionId ?? this.sessionId,
        stepNumber: stepNumber ?? this.stepNumber,
        instruction: instruction ?? this.instruction,
        verified: verified ?? this.verified,
        createdAt: createdAt ?? this.createdAt,
      );

  // ---------------------------------------------------------------------------
  // Equality
  // ---------------------------------------------------------------------------

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is SessionStep &&
          runtimeType == other.runtimeType &&
          id == other.id &&
          sessionId == other.sessionId &&
          stepNumber == other.stepNumber &&
          instruction == other.instruction &&
          verified == other.verified &&
          createdAt == other.createdAt;

  @override
  int get hashCode => Object.hash(
        id,
        sessionId,
        stepNumber,
        instruction,
        verified,
        createdAt,
      );

  @override
  String toString() =>
      'SessionStep(id: $id, sessionId: $sessionId, stepNumber: $stepNumber, '
      'instruction: $instruction, verified: $verified, createdAt: $createdAt)';
}
