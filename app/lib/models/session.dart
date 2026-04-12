/// Status of a repair session.
enum SessionStatus {
  active,
  completed,
  abandoned;

  /// Converts the string value stored in Supabase to [SessionStatus].
  static SessionStatus fromJson(String value) => switch (value) {
        'active' => SessionStatus.active,
        'completed' => SessionStatus.completed,
        'abandoned' => SessionStatus.abandoned,
        _ => throw ArgumentError('Unknown SessionStatus: $value'),
      };

  String toJson() => name;
}

/// Represents a single repair session tied to a user and optionally a device.
///
/// Maps directly to the `sessions` table in Supabase.
class Session {
  final String id;
  final String userId;

  /// Nullable — a session may exist before a device is identified.
  final String? deviceId;

  final SessionStatus status;

  /// Free-text description of what went wrong, filled in during triage.
  final String? problemDetected;

  /// Free-text description of the fix that was applied.
  final String? solutionApplied;

  /// Whether the repair was ultimately successful. Null until the session ends.
  final bool? success;

  /// Total elapsed time in seconds. Null while the session is still active.
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
        status: SessionStatus.fromJson(json['status'] as String),
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

  bool get isActive => status == SessionStatus.active;
  bool get isCompleted => status == SessionStatus.completed;

  /// Human-readable duration string.
  ///
  /// Examples: "2h 15m", "3m 42s", "< 1m"
  String get durationFormatted {
    final secs = durationSeconds;
    if (secs == null || secs <= 0) return '< 1m';

    final h = secs ~/ 3600;
    final m = (secs % 3600) ~/ 60;
    final s = secs % 60;

    if (h > 0) return '${h}h ${m}m';
    if (m > 0) return '${m}m ${s}s';
    return '< 1m';
  }

  // ---------------------------------------------------------------------------
  // copyWith
  // ---------------------------------------------------------------------------

  Session copyWith({
    String? id,
    String? userId,
    // Use a sentinel to distinguish "pass null explicitly" from "keep current".
    Object? deviceId = _keep,
    SessionStatus? status,
    Object? problemDetected = _keep,
    Object? solutionApplied = _keep,
    Object? success = _keep,
    Object? durationSeconds = _keep,
    DateTime? createdAt,
  }) =>
      Session(
        id: id ?? this.id,
        userId: userId ?? this.userId,
        deviceId: deviceId == _keep ? this.deviceId : deviceId as String?,
        status: status ?? this.status,
        problemDetected: problemDetected == _keep
            ? this.problemDetected
            : problemDetected as String?,
        solutionApplied: solutionApplied == _keep
            ? this.solutionApplied
            : solutionApplied as String?,
        success: success == _keep ? this.success : success as bool?,
        durationSeconds: durationSeconds == _keep
            ? this.durationSeconds
            : durationSeconds as int?,
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
      'status: $status, success: $success, '
      'durationSeconds: $durationSeconds, createdAt: $createdAt)';
}

/// Private sentinel used by [Session.copyWith] to distinguish an explicit
/// `null` argument from a missing (keep-current) argument.
const Object _keep = Object();
