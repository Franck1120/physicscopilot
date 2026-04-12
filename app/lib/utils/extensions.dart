extension DurationFormatting on Duration {
  String formatShort() {
    final totalSeconds = inSeconds;

    if (totalSeconds == 0) return '< 1m';

    if (totalSeconds < 60) return '${totalSeconds}s';

    if (totalSeconds < 3600) {
      final minutes = totalSeconds ~/ 60;
      final seconds = totalSeconds % 60;
      if (seconds == 0) return '${minutes}m';
      return '${minutes}m ${seconds}s';
    }

    final hours = totalSeconds ~/ 3600;
    final minutes = (totalSeconds % 3600) ~/ 60;
    if (minutes == 0) return '${hours}h';
    return '${hours}h ${minutes}m';
  }
}

extension DateTimeFormatting on DateTime {
  static const List<String> _weekdays = [
    'lun',
    'mar',
    'mer',
    'gio',
    'ven',
    'sab',
    'dom',
  ];

  static const List<String> _months = [
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

  String _paddedHHmm() {
    final hh = hour.toString().padLeft(2, '0');
    final mm = minute.toString().padLeft(2, '0');
    return '$hh:$mm';
  }

  String formatRelative() {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final thisDate = DateTime(year, month, day);
    final difference = today.difference(thisDate).inDays;

    if (difference == 0) return 'oggi ${_paddedHHmm()}';
    if (difference == 1) return 'ieri ${_paddedHHmm()}';
    if (difference < 7) {
      // weekday: 1=Monday … 7=Sunday; array is 0-indexed lun=0 … dom=6
      final dayName = _weekdays[weekday - 1];
      return '$dayName ${_paddedHHmm()}';
    }

    final monthName = _months[month - 1];
    if (year != now.year) return '$day $monthName $year';
    return '$day $monthName';
  }

  String formatTimestamp() {
    final monthName = _months[month - 1];
    return '$day $monthName $year, ${_paddedHHmm()}';
  }
}

extension StringUtils on String {
  String capitalize() {
    if (isEmpty) return this;
    return '${this[0].toUpperCase()}${substring(1)}';
  }

  String truncate(int maxLength, {String ellipsis = '…'}) {
    if (length <= maxLength) return this;
    return '${substring(0, maxLength)}$ellipsis';
  }

  bool isBlank() => trim().isEmpty;
}

extension IntDuration on int {
  Duration get asDuration => Duration(seconds: this);

  String formatAsTimer() {
    final totalSeconds = this < 0 ? 0 : this;
    final hours = totalSeconds ~/ 3600;
    final minutes = (totalSeconds % 3600) ~/ 60;
    final seconds = totalSeconds % 60;

    final mm = minutes.toString().padLeft(2, '0');
    final ss = seconds.toString().padLeft(2, '0');

    if (hours > 0) {
      final hh = hours.toString().padLeft(2, '0');
      return '$hh:$mm:$ss';
    }
    return '$mm:$ss';
  }
}
