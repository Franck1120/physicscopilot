/// Status of a repair session.
enum SessionStatus {
  active,
  completed,
  abandoned;

  /// Parses a raw DB string. Falls back to [active] for unknown values.
  static SessionStatus fromString(String s) => switch (s) {
        'completed' => SessionStatus.completed,
        'abandoned' => SessionStatus.abandoned,
        _ => SessionStatus.active,
      };

  /// Localised label shown in the UI.
  String get label => switch (this) {
        SessionStatus.active => 'Attiva',
        SessionStatus.completed => 'Completata',
        SessionStatus.abandoned => 'Abbandonata',
      };
}

// ---------------------------------------------------------------------------

/// A single step inside a [Session], mapped to the `session_steps` table.
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
    this.verified = false,
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
        verified: (json['verified'] as bool?) ?? false,
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
  int get hashCode =>
      Object.hash(id, sessionId, stepNumber, instruction, verified, createdAt);

  @override
  String toString() =>
      'SessionStep(id: $id, sessionId: $sessionId, stepNumber: $stepNumber, '
      'verified: $verified)';
}

// ---------------------------------------------------------------------------

/// A repair session, mapped to the `sessions` table in Supabase.
///
/// [steps] is populated client-side from a join on `session_steps` and is
/// intentionally excluded from [toJson] (never written back to the `sessions`
/// row).
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
  final List<SessionStep> steps;

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
    this.steps = const [],
  });

  // ---------------------------------------------------------------------------
  // Serialization
  // ---------------------------------------------------------------------------

  factory Session.fromJson(Map<String, dynamic> json) => Session(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        deviceId: json['device_id'] as String?,
        status: SessionStatus.fromString(
          (json['status'] as String?) ?? 'active',
        ),
        problemDetected: json['problem_detected'] as String?,
        solutionApplied: json['solution_applied'] as String?,
        success: json['success'] as bool?,
        durationSeconds: json['duration_seconds'] as int?,
        createdAt: DateTime.parse(json['created_at'] as String),
        // steps are populated separately — not stored in the sessions row
        steps: const [],
      );

  /// Serializes to the `sessions` DB row format.
  /// [steps] is intentionally excluded.
  Map<String, dynamic> toJson() => {
        'id': id,
        'user_id': userId,
        if (deviceId != null) 'device_id': deviceId,
        'status': status.name,
        if (problemDetected != null) 'problem_detected': problemDetected,
        if (solutionApplied != null) 'solution_applied': solutionApplied,
        if (success != null) 'success': success,
        if (durationSeconds != null) 'duration_seconds': durationSeconds,
        'created_at': createdAt.toIso8601String(),
      };

  // ---------------------------------------------------------------------------
  // Convenience
  // ---------------------------------------------------------------------------

  /// Returns the session duration, or null if [durationSeconds] is not set.
  Duration? get duration =>
      durationSeconds != null ? Duration(seconds: durationSeconds!) : null;

  // ---------------------------------------------------------------------------
  // copyWith
  // ---------------------------------------------------------------------------

  Session copyWith({
    String? id,
    String? userId,
    // Use a sentinel to distinguish "clear deviceId" from "leave unchanged".
    Object? deviceId = _unset,
    SessionStatus? status,
    Object? problemDetected = _unset,
    Object? solutionApplied = _unset,
    Object? success = _unset,
    Object? durationSeconds = _unset,
    DateTime? createdAt,
    List<SessionStep>? steps,
  }) =>
      Session(
        id: id ?? this.id,
        userId: userId ?? this.userId,
        deviceId: deviceId == _unset ? this.deviceId : deviceId as String?,
        status: status ?? this.status,
        problemDetected: problemDetected == _unset
            ? this.problemDetected
            : problemDetected as String?,
        solutionApplied: solutionApplied == _unset
            ? this.solutionApplied
            : solutionApplied as String?,
        success: success == _unset ? this.success : success as bool?,
        durationSeconds: durationSeconds == _unset
            ? this.durationSeconds
            : durationSeconds as int?,
        createdAt: createdAt ?? this.createdAt,
        steps: steps ?? this.steps,
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
      'status: $status, success: $success, createdAt: $createdAt)';
}

// Sentinel used in copyWith to detect "not provided" for nullable fields.
const Object _unset = Object();
