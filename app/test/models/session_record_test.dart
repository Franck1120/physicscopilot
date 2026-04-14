import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/models/session_record.dart';

void main() {
  group('SessionRecord', () {
    test('constructor creates with all required fields', () {
      final record = SessionRecord(
        id: 'rec-001',
        date: DateTime(2026, 4, 10),
        equipmentName: 'HP LaserJet 1020',
        summary: '',
        problemDescription: 'Paper jam at tray 2',
        status: SessionStatus.resolved,
        duration: const Duration(minutes: 8),
      );
      expect(record.id, 'rec-001');
      expect(record.equipmentName, 'HP LaserJet 1020');
      expect(record.problemDescription, 'Paper jam at tray 2');
      expect(record.status, SessionStatus.resolved);
      expect(record.duration.inMinutes, 8);
    });

    test('SessionStatus enum has exactly resolved and unresolved values', () {
      expect(SessionStatus.values.length, 2);
      expect(
        SessionStatus.values,
        containsAll([SessionStatus.resolved, SessionStatus.unresolved]),
      );
    });

    test('date field is preserved with full precision', () {
      final date = DateTime(2026, 4, 10, 14, 30, 45);
      final record = SessionRecord(
        id: 'id',
        date: date,
        equipmentName: 'Printer',
        summary: '',
        problemDescription: 'Issue',
        status: SessionStatus.unresolved,
        duration: Duration.zero,
      );
      expect(record.date, date);
      expect(record.date.hour, 14);
      expect(record.date.minute, 30);
    });

    test('duration is stored and accessible in multiple units', () {
      final record = SessionRecord(
        id: 'id',
        date: DateTime.now(),
        equipmentName: 'Printer',
        summary: '',
        problemDescription: 'Issue',
        status: SessionStatus.resolved,
        duration: const Duration(minutes: 5, seconds: 30),
      );
      expect(record.duration.inMinutes, 5);
      expect(record.duration.inSeconds, 330);
    });

    test('unresolved status is distinct from resolved', () {
      final resolved = SessionRecord(
        id: 'r1',
        date: DateTime.now(),
        equipmentName: 'P',
        summary: '',
        problemDescription: 'D',
        status: SessionStatus.resolved,
        duration: Duration.zero,
      );
      final unresolved = SessionRecord(
        id: 'r2',
        date: DateTime.now(),
        equipmentName: 'P',
        summary: '',
        problemDescription: 'D',
        status: SessionStatus.unresolved,
        duration: Duration.zero,
      );
      expect(resolved.status, isNot(equals(unresolved.status)));
    });
  });
}
