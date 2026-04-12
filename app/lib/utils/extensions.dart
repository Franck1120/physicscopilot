// Dart extension methods — utility common types

/// Duration formatting helpers.
extension DurationFormatting on Duration {
  /// Clock-style representation: `m:ss` or `h:mm:ss` when ≥ 1 hour.
  ///
  /// Examples: `2:05`, `1:03:07`
  String get formatted {
    final h = inHours;
    final m = inMinutes.remainder(60);
    final s = inSeconds.remainder(60);

    final mm = m.toString().padLeft(2, '0');
    final ss = s.toString().padLeft(2, '0');

    if (h > 0) {
      return '$h:$mm:$ss';
    }
    // Drop the leading zero on minutes for a natural clock look (2:05 not 02:05).
    return '$m:$ss';
  }

  /// Human-readable Italian approximation.
  ///
  /// Examples: `45 sec`, `2 min`, `1h 5min`
  String get humanReadable {
    final h = inHours;
    final m = inMinutes.remainder(60);
    final s = inSeconds.remainder(60);

    if (h > 0) {
      return m > 0 ? '${h}h ${m}min' : '${h}h';
    }
    if (m > 0) return '$m min';
    return '$s sec';
  }
}

// ---------------------------------------------------------------------------

const _kMonthsShort = [
  'gen',
  'feb',
  'mar',
  'apr',
  'mag',
  'giu',
  'lug',
  'ago',
  'set',
  'ott',
  'nov',
  'dic',
];

/// DateTime formatting helpers (Italian locale).
extension DateTimeFormatting on DateTime {
  /// Day, abbreviated month and year in Italian: `12 apr 2026`.
  String get formattedDate {
    final month = _kMonthsShort[this.month - 1];
    return '$day $month $year';
  }

  /// 24-hour clock: `14:35`.
  String get formattedTime {
    final hh = hour.toString().padLeft(2, '0');
    final mm = minute.toString().padLeft(2, '0');
    return '$hh:$mm';
  }

  /// Relative label in Italian, falling back to [formattedDate].
  ///
  /// - Same calendar day → `oggi`
  /// - Previous calendar day → `ieri`
  /// - 2–6 days ago → `3 giorni fa`
  /// - Older → `12 apr 2026`
  String get relativeTime {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final thisDay = DateTime(year, month, day);
    final diff = today.difference(thisDay).inDays;

    if (diff == 0) return 'oggi';
    if (diff == 1) return 'ieri';
    if (diff >= 2 && diff <= 6) return '$diff giorni fa';
    return formattedDate;
  }
}

// ---------------------------------------------------------------------------

/// String utility helpers.
extension StringUtilities on String {
  /// Returns the string with its first character uppercased.
  ///
  /// Returns an empty string unchanged.
  String get capitalize {
    if (isEmpty) return this;
    return this[0].toUpperCase() + substring(1);
  }

  /// Returns `true` when the string starts with a recognised URL scheme.
  bool get isValidUrl {
    return startsWith('http://') ||
        startsWith('https://') ||
        startsWith('ws://') ||
        startsWith('wss://');
  }
}

// ---------------------------------------------------------------------------

/// Convenience constructors on [int] for time-based [Duration] values.
extension IntDuration on int {
  /// `42.seconds` → `Duration(seconds: 42)`
  Duration get seconds => Duration(seconds: this);

  /// `5.minutes` → `Duration(minutes: 5)`
  Duration get minutes => Duration(minutes: this);
}
