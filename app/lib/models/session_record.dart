// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'dart:convert';

/// Outcome of a completed session stored in local history.
enum SessionStatus {
  /// The problem was resolved successfully.
  resolved,

  /// The session ended without a confirmed resolution.
  unresolved,
}

/// A completed session persisted to local history (SharedPreferences).
///
/// Serialised as JSON and stored as a capped list via [SessionHistoryNotifier].
class SessionRecord {
  final String id;
  final DateTime date;
  final String equipmentName;
  final String problemDescription;

  /// AI-generated summary of what was found/done during the session.
  final String summary;
  final SessionStatus status;
  final Duration duration;

  const SessionRecord({
    required this.id,
    required this.date,
    required this.equipmentName,
    required this.problemDescription,
    required this.summary,
    required this.status,
    required this.duration,
  });

  /// Serialises this record to a JSON-compatible map.
  Map<String, dynamic> toJson() => {
        'id': id,
        'date': date.toIso8601String(),
        'equipmentName': equipmentName,
        'problemDescription': problemDescription,
        'summary': summary,
        'status': status.name,
        'durationSeconds': duration.inSeconds,
      };

  /// Parses a record from a JSON map produced by [toJson].
  factory SessionRecord.fromJson(Map<String, dynamic> json) => SessionRecord(
        id: json['id'] as String,
        date: DateTime.parse(json['date'] as String),
        equipmentName: (json['equipmentName'] as String?) ?? '',
        problemDescription: (json['problemDescription'] as String?) ?? '',
        summary: (json['summary'] as String?) ?? '',
        status: (json['status'] as String?) == 'resolved'
            ? SessionStatus.resolved
            : SessionStatus.unresolved,
        duration: Duration(seconds: (json['durationSeconds'] as int?) ?? 0),
      );

  /// Encodes a list of records to a JSON string for SharedPreferences storage.
  static String encodeList(List<SessionRecord> records) =>
      jsonEncode(records.map((r) => r.toJson()).toList());

  /// Decodes a JSON string produced by [encodeList] back to a list of records.
  static List<SessionRecord> decodeList(String raw) {
    final list = jsonDecode(raw) as List<dynamic>;
    return list
        .map((e) => SessionRecord.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
