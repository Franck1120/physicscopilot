import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/models/session_record.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/providers/session_history_provider.dart';

// ── Helpers ───────────────────────────────────────────────────────────────────

ProviderContainer makeContainer(SharedPreferences prefs) => ProviderContainer(
      overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
    );

SessionRecord _makeRecord({
  String id = 'id-1',
  String equipmentName = 'Printer X',
  String problemDescription = 'Paper jam',
  SessionStatus status = SessionStatus.resolved,
}) =>
    SessionRecord(
      id: id,
      date: DateTime(2024, 1, 1),
      equipmentName: equipmentName,
      problemDescription: problemDescription,
      summary: 'Cleared the jam.',
      status: status,
      duration: const Duration(minutes: 10),
    );

// ── Tests ─────────────────────────────────────────────────────────────────────

void main() {
  group('SessionHistoryNotifier', () {
    setUp(() {
      SharedPreferences.setMockInitialValues({});
    });

    test('initial state is empty when prefs are empty', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = makeContainer(prefs);
      addTearDown(container.dispose);

      expect(container.read(sessionHistoryProvider), isEmpty);
    });

    test('add(record) → record appears at state[0]', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = makeContainer(prefs);
      addTearDown(container.dispose);

      final record = _makeRecord();
      await container.read(sessionHistoryProvider.notifier).add(record);

      final state = container.read(sessionHistoryProvider);
      expect(state, hasLength(1));
      expect(state.first.id, record.id);
    });

    test('add multiple records → ordered newest-first', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = makeContainer(prefs);
      addTearDown(container.dispose);

      final older = _makeRecord(id: 'old');
      final newer = _makeRecord(id: 'new');
      await container.read(sessionHistoryProvider.notifier).add(older);
      await container.read(sessionHistoryProvider.notifier).add(newer);

      final state = container.read(sessionHistoryProvider);
      expect(state.first.id, 'new');
      expect(state.last.id, 'old');
    });

    test('remove(id) → record removed from state', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = makeContainer(prefs);
      addTearDown(container.dispose);

      final r1 = _makeRecord(id: 'id-1');
      final r2 = _makeRecord(id: 'id-2');
      await container.read(sessionHistoryProvider.notifier).add(r1);
      await container.read(sessionHistoryProvider.notifier).add(r2);

      await container.read(sessionHistoryProvider.notifier).remove('id-1');

      final state = container.read(sessionHistoryProvider);
      expect(state, hasLength(1));
      expect(state.first.id, 'id-2');
    });

    test('clearAll() → state is empty and prefs key removed', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = makeContainer(prefs);
      addTearDown(container.dispose);

      await container.read(sessionHistoryProvider.notifier).add(_makeRecord());
      await container.read(sessionHistoryProvider.notifier).clearAll();

      expect(container.read(sessionHistoryProvider), isEmpty);
      expect(prefs.containsKey('session_history'), isFalse);
    });

    test('cap at 50 records — adding 51 keeps only the 50 most recent', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = makeContainer(prefs);
      addTearDown(container.dispose);

      for (var i = 0; i < 51; i++) {
        await container
            .read(sessionHistoryProvider.notifier)
            .add(_makeRecord(id: 'id-$i'));
      }

      expect(container.read(sessionHistoryProvider), hasLength(50));
      // The oldest record (id-0) should have been dropped
      expect(
        container.read(sessionHistoryProvider).any((r) => r.id == 'id-0'),
        isFalse,
      );
    });

    test('persistence: after add, recreating container reloads record from prefs', () async {
      final prefs = await SharedPreferences.getInstance();
      final container1 = makeContainer(prefs);
      final record = _makeRecord(id: 'persistent-id');
      await container1.read(sessionHistoryProvider.notifier).add(record);
      container1.dispose();

      // Recreate container with the same prefs instance (data is persisted)
      final container2 = makeContainer(prefs);
      addTearDown(container2.dispose);

      final state = container2.read(sessionHistoryProvider);
      expect(state, hasLength(1));
      expect(state.first.id, 'persistent-id');
    });
  });
}
