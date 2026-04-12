// Unit tests for milestone_provider.dart
//
// Verifies that earnedMilestonesProvider and nextMilestoneProvider return the
// correct values for various session-count scenarios.
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:physicscopilot/providers/prefs_provider.dart'
    show sharedPrefsProvider;
import 'package:physicscopilot/providers/session_history_provider.dart';
import 'package:physicscopilot/providers/milestone_provider.dart';
import 'package:physicscopilot/models/session_record.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

List<SessionRecord> _records(int n) => List.generate(
      n,
      (i) => SessionRecord(
        id: 'id-$i',
        date: DateTime(2026, 1, i + 1),
        equipmentName: 'Test',
        problemDescription: '',
        summary: '',
        status: SessionStatus.resolved,
        duration: Duration.zero,
      ),
    );

ProviderContainer _buildContainer(int sessionCount) {
  final encoded = SessionRecord.encodeList(_records(sessionCount));
  SharedPreferences.setMockInitialValues({'session_history': encoded});
  // We create a fresh container per test; shared_preferences mock values
  // are set synchronously, so getInstance() inside the notifier returns them.
  return ProviderContainer(
    overrides: [],
  );
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('earnedMilestonesProvider', () {
    test('returns empty list when 0 sessions', () async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          sessionHistoryProvider.overrideWith(
            (ref) => SessionHistoryNotifier(prefs),
          ),
        ],
      );
      addTearDown(container.dispose);

      expect(container.read(earnedMilestonesProvider), isEmpty);
    });

    test('returns [5] when exactly 5 sessions', () async {
      final encoded = SessionRecord.encodeList(_records(5));
      SharedPreferences.setMockInitialValues({'session_history': encoded});
      final prefs = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          sessionHistoryProvider.overrideWith(
            (ref) => SessionHistoryNotifier(prefs),
          ),
        ],
      );
      addTearDown(container.dispose);

      expect(container.read(earnedMilestonesProvider), equals([5]));
    });

    test('returns [5, 10] when 10 sessions', () async {
      final encoded = SessionRecord.encodeList(_records(10));
      SharedPreferences.setMockInitialValues({'session_history': encoded});
      final prefs = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          sessionHistoryProvider.overrideWith(
            (ref) => SessionHistoryNotifier(prefs),
          ),
        ],
      );
      addTearDown(container.dispose);

      expect(container.read(earnedMilestonesProvider), equals([5, 10]));
    });

    test('returns all milestones when 50+ sessions', () async {
      final encoded = SessionRecord.encodeList(_records(50));
      SharedPreferences.setMockInitialValues({'session_history': encoded});
      final prefs = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          sessionHistoryProvider.overrideWith(
            (ref) => SessionHistoryNotifier(prefs),
          ),
        ],
      );
      addTearDown(container.dispose);

      expect(
        container.read(earnedMilestonesProvider),
        equals([5, 10, 25, 50]),
      );
    });
  });

  group('nextMilestoneProvider', () {
    test('returns 5 when 0 sessions', () async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          sessionHistoryProvider.overrideWith(
            (ref) => SessionHistoryNotifier(prefs),
          ),
        ],
      );
      addTearDown(container.dispose);

      expect(container.read(nextMilestoneProvider), equals(5));
    });

    test('returns 10 when 5 sessions', () async {
      final encoded = SessionRecord.encodeList(_records(5));
      SharedPreferences.setMockInitialValues({'session_history': encoded});
      final prefs = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          sessionHistoryProvider.overrideWith(
            (ref) => SessionHistoryNotifier(prefs),
          ),
        ],
      );
      addTearDown(container.dispose);

      expect(container.read(nextMilestoneProvider), equals(10));
    });

    test('returns null when all milestones earned (50+ sessions)', () async {
      final encoded = SessionRecord.encodeList(_records(50));
      SharedPreferences.setMockInitialValues({'session_history': encoded});
      final prefs = await SharedPreferences.getInstance();
      final container = ProviderContainer(
        overrides: [
          sharedPrefsProvider.overrideWithValue(prefs),
          sessionHistoryProvider.overrideWith(
            (ref) => SessionHistoryNotifier(prefs),
          ),
        ],
      );
      addTearDown(container.dispose);

      expect(container.read(nextMilestoneProvider), isNull);
    });
  });
}
