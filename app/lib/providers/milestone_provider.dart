// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'session_history_provider.dart';

/// Session-count thresholds that earn a milestone badge.
const List<int> kMilestoneThresholds = [5, 10, 25, 50];

/// Returns the list of milestone thresholds that the user has already reached,
/// based on the total number of completed sessions in [sessionHistoryProvider].
///
/// Example: if the user has completed 12 sessions the result is `[5, 10]`.
final earnedMilestonesProvider = Provider<List<int>>((ref) {
  final sessions = ref.watch(sessionHistoryProvider);
  final count = sessions.length;
  return kMilestoneThresholds.where((t) => count >= t).toList();
});

/// Returns the next milestone the user has not yet reached, or `null` if all
/// milestones have been earned.
final nextMilestoneProvider = Provider<int?>((ref) {
  final sessions = ref.watch(sessionHistoryProvider);
  final count = sessions.length;
  for (final t in kMilestoneThresholds) {
    if (count < t) return t;
  }
  return null;
});
