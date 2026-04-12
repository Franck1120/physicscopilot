// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../models/session_record.dart';
import 'prefs_provider.dart';

const _kHistoryKey = 'session_history';
const _kMaxRecords = 50;

/// Persists and exposes the list of completed sessions (newest-first).
///
/// Backed by SharedPreferences as a JSON array.
class SessionHistoryNotifier extends StateNotifier<List<SessionRecord>> {
  SessionHistoryNotifier(this._prefs) : super(_load(_prefs));

  final SharedPreferences _prefs;

  static List<SessionRecord> _load(SharedPreferences prefs) {
    final raw = prefs.getString(_kHistoryKey);
    if (raw == null) return [];
    try {
      return SessionRecord.decodeList(raw);
    } catch (_) {
      return [];
    }
  }

  /// Prepends [record] to the list and persists.
  Future<void> add(SessionRecord record) async {
    final updated = [record, ...state];
    final capped = updated.length > _kMaxRecords
        ? updated.sublist(0, _kMaxRecords)
        : updated;
    state = capped;
    await _prefs.setString(_kHistoryKey, SessionRecord.encodeList(capped));
  }

  /// Removes a single record by [id].
  Future<void> remove(String id) async {
    final updated = state.where((r) => r.id != id).toList();
    state = updated;
    await _prefs.setString(_kHistoryKey, SessionRecord.encodeList(updated));
  }

  /// Wipes all history.
  Future<void> clearAll() async {
    state = [];
    await _prefs.remove(_kHistoryKey);
  }
}

/// Provides the [SessionHistoryNotifier] and the persisted list of sessions.
///
/// Backed by SharedPreferences; the list is capped at [_kMaxRecords] entries.
final sessionHistoryProvider =
    StateNotifierProvider<SessionHistoryNotifier, List<SessionRecord>>((ref) {
  final prefs = ref.watch(sharedPrefsProvider);
  return SessionHistoryNotifier(prefs);
});
