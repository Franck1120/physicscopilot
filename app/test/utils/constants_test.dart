import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/utils/constants.dart';

void main() {
  group('AppConstants', () {
    test('wsBaseUrl starts with wss://', () {
      expect(AppConstants.wsBaseUrl, startsWith('wss://'));
    });

    test('apiBaseUrl starts with https://', () {
      expect(AppConstants.apiBaseUrl, startsWith('https://'));
    });

    test('wsBaseUrl is not empty', () {
      expect(AppConstants.wsBaseUrl, isNotEmpty);
    });

    test('apiBaseUrl is not empty', () {
      expect(AppConstants.apiBaseUrl, isNotEmpty);
    });

    test('wsBaseUrl and apiBaseUrl share the same host', () {
      final wsHost = AppConstants.wsBaseUrl.replaceFirst('wss://', '');
      final apiHost = AppConstants.apiBaseUrl.replaceFirst('https://', '');
      expect(wsHost, equals(apiHost));
    });

    test('wsBaseUrl contains a valid hostname (no spaces)', () {
      final host = AppConstants.wsBaseUrl.replaceFirst('wss://', '');
      expect(host.contains(' '), isFalse);
    });

    test('apiBaseUrl contains a valid hostname (no spaces)', () {
      final host = AppConstants.apiBaseUrl.replaceFirst('https://', '');
      expect(host.contains(' '), isFalse);
    });

    test('wsBaseUrl and apiBaseUrl have different schemes', () {
      expect(AppConstants.wsBaseUrl, isNot(equals(AppConstants.apiBaseUrl)));
    });
  });
}
