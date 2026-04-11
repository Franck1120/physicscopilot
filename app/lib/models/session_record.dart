enum SessionStatus { resolved, unresolved }

class SessionRecord {
  final String id;
  final DateTime date;
  final String printerName;
  final String problemDescription;
  final SessionStatus status;
  final Duration duration;

  const SessionRecord({
    required this.id,
    required this.date,
    required this.printerName,
    required this.problemDescription,
    required this.status,
    required this.duration,
  });
}
