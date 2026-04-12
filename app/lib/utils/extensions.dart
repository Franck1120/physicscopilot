// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

// Dart extension methods for common types.
// Pure Dart utilities (no Flutter dependencies).

// ── Duration extension ──────────────────────────────────────────────────────

/// Extension methods on [Duration].
extension DurationX on Duration {
  /// Formats duration in Italian.
  ///
  /// Examples:
  /// - `5 minutes` → `"5 min"`
  /// - `90 minutes` → `"1h 30min"`
  /// - `120 minutes` → `"2h"`
  String get formatted {
    final minutes = inMinutes;
    if (minutes < 60) return '$minutes min';

    final hours = inHours;
    final remaining = minutes % 60;
    return remaining == 0 ? '${hours}h' : '${hours}h ${remaining}min';
  }
}

// ── DateTime extension ──────────────────────────────────────────────────────

/// Extension methods on [DateTime].
extension DateTimeX on DateTime {
  /// Formats date in short Italian format.
  ///
  /// Example: `"12 apr 2026"`
  String get shortDate {
    const months = [
      'gen', 'feb', 'mar', 'apr', 'mag', 'giu',
      'lug', 'ago', 'set', 'ott', 'nov', 'dic',
    ];
    return '$day ${months[month - 1]} $year';
  }

  /// Returns relative label in Italian based on how many days ago.
  ///
  /// Examples:
  /// - Today → `"Oggi"`
  /// - Yesterday → `"Ieri"`
  /// - 3 days ago → `"3 giorni fa"`
  /// - Before last 7 days → formatted date like `"12 apr 2026"`
  String get relativeLabel {
    final now = DateTime.now();

    // Compare by date only (ignore time)
    final isToday = year == now.year &&
        month == now.month &&
        day == now.day;

    if (isToday) return 'Oggi';

    final yesterday = DateTime(now.year, now.month, now.day - 1);
    final isYesterday = year == yesterday.year &&
        month == yesterday.month &&
        day == yesterday.day;

    if (isYesterday) return 'Ieri';

    // Calculate days between (from this date to now)
    final daysDiff = now.difference(this).inDays;
    if (daysDiff > 0 && daysDiff < 7) {
      return '$daysDiff giorni fa';
    }

    // Fallback: return short date format
    return shortDate;
  }
}

// ── String extension ────────────────────────────────────────────────────────

/// Extension methods on [String].
extension StringX on String {
  /// Extracts first letter(s) from each word, max 2 characters.
  ///
  /// Examples:
  /// - `"John Doe"` → `"JD"`
  /// - `"Franck"` → `"F"`
  /// - `"Alice Bob Charlie"` → `"AB"`
  /// - `""` → `"?"`
  String get initials {
    if (isEmpty) return '?';

    return split(' ')
        .where((word) => word.isNotEmpty)
        .take(2)
        .map((word) => word[0].toUpperCase())
        .join();
  }
}
