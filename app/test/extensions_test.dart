import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/utils/extensions.dart';

void main() {
  group('DurationX.formatted', () {
    test('formats minutes under 60', () {
      expect(const Duration(minutes: 5).formatted, equals('5 min'));
      expect(const Duration(minutes: 59).formatted, equals('59 min'));
      expect(const Duration(minutes: 1).formatted, equals('1 min'));
    });

    test('formats hours without remaining minutes', () {
      expect(const Duration(hours: 1).formatted, equals('1h'));
      expect(const Duration(hours: 2).formatted, equals('2h'));
    });

    test('formats hours with remaining minutes', () {
      expect(const Duration(hours: 1, minutes: 30).formatted, equals('1h 30min'));
      expect(const Duration(hours: 2, minutes: 45).formatted, equals('2h 45min'));
    });

    test('formats zero duration', () {
      expect(const Duration(minutes: 0).formatted, equals('0 min'));
    });
  });

  group('DateTimeX.shortDate', () {
    test('formats date correctly in Italian', () {
      final date = DateTime(2026, 4, 12);
      expect(date.shortDate, equals('12 apr 2026'));
    });

    test('handles different months', () {
      expect(DateTime(2026, 1, 5).shortDate, equals('5 gen 2026'));
      expect(DateTime(2026, 12, 25).shortDate, equals('25 dic 2026'));
      expect(DateTime(2026, 6, 15).shortDate, equals('15 giu 2026'));
    });

    test('handles day 1 and last day', () {
      expect(DateTime(2026, 3, 1).shortDate, equals('1 mar 2026'));
      expect(DateTime(2026, 2, 28).shortDate, equals('28 feb 2026'));
    });
  });

  group('DateTimeX.relativeLabel', () {
    test('returns "Oggi" for today', () {
      final today = DateTime.now();
      expect(today.relativeLabel, equals('Oggi'));
    });

    test('returns "Ieri" for yesterday', () {
      final yesterday = DateTime.now().subtract(const Duration(days: 1));
      expect(yesterday.relativeLabel, equals('Ieri'));
    });

    test('returns "N giorni fa" for days within last 7 days', () {
      final threeDaysAgo = DateTime.now().subtract(const Duration(days: 3));
      expect(threeDaysAgo.relativeLabel, equals('3 giorni fa'));

      final sixDaysAgo = DateTime.now().subtract(const Duration(days: 6));
      expect(sixDaysAgo.relativeLabel, equals('6 giorni fa'));
    });

    test('returns short date format for older dates', () {
      final tenDaysAgo = DateTime.now().subtract(const Duration(days: 10));
      // Should return shortDate format, not relative
      expect(tenDaysAgo.relativeLabel, equals(tenDaysAgo.shortDate));
    });

    test('ignores time component when comparing dates', () {
      final now = DateTime.now();
      final todayAtMidnight = DateTime(now.year, now.month, now.day);
      expect(todayAtMidnight.relativeLabel, equals('Oggi'));

      final yesterdayAtNoon =
          DateTime(now.year, now.month, now.day - 1, 12, 30, 45);
      expect(yesterdayAtNoon.relativeLabel, equals('Ieri'));
    });
  });

  group('StringX.initials', () {
    test('extracts initials from two-word names', () {
      expect('John Doe'.initials, equals('JD'));
      expect('Alice Smith'.initials, equals('AS'));
    });

    test('returns single initial for one-word names', () {
      expect('Franck'.initials, equals('F'));
      expect('Bob'.initials, equals('B'));
    });

    test('takes only first 2 initials from longer names', () {
      expect('Alice Bob Charlie'.initials, equals('AB'));
      expect('John Jacob Jingleheimer Schmidt'.initials, equals('JJ'));
    });

    test('returns "?" for empty string', () {
      expect(''.initials, equals('?'));
    });

    test('handles names with multiple spaces', () {
      expect('John  Doe'.initials, equals('JD'));
      expect('  Alice  Smith  '.initials, equals('AS'));
    });

    test('converts to uppercase', () {
      expect('john doe'.initials, equals('JD'));
      expect('alice smith'.initials, equals('AS'));
    });

    test('handles single-letter words', () {
      expect('A B C'.initials, equals('AB'));
      expect('X'.initials, equals('X'));
    });
  });
}
