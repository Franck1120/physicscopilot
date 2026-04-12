import 'dart:convert';

enum SessionStatus { resolved, unresolved }

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

  Map<String, dynamic> toJson() => {
        'id': id,
        'date': date.toIso8601String(),
        'equipmentName': equipmentName,
        'problemDescription': problemDescription,
        'summary': summary,
        'status': status.name,
        'durationSeconds': duration.inSeconds,
      };

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

  static String encodeList(List<SessionRecord> records) =>
      jsonEncode(records.map((r) => r.toJson()).toList());

  static List<SessionRecord> decodeList(String raw) {
    final list = jsonDecode(raw) as List<dynamic>;
    return list
        .map((e) => SessionRecord.fromJson(e as Map<String, dynamic>))
        .toList();
  }
}
