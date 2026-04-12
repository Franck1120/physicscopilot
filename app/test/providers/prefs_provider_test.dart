// Unit tests for sharedPrefsProvider.
//
// The provider is intentionally declared to throw UnimplementedError unless
// overridden — these tests verify that contract and that the override works.
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:physicscopilot/providers/prefs_provider.dart';

void main() {
  group('sharedPrefsProvider', () {
    test('reading without override throws UnimplementedError', () {
      final container = ProviderContainer();
      addTearDown(container.dispose);

      expect(
        () => container.read(sharedPrefsProvider),
        throwsA(isA<UnimplementedError>()),
      );
    });

    test('with override returns the mocked SharedPreferences instance',
        () async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      final container = ProviderContainer(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
      );
      addTearDown(container.dispose);

      final result = container.read(sharedPrefsProvider);
      expect(result, same(prefs));
    });

    test('override with data — stored string value is accessible', () async {
      SharedPreferences.setMockInitialValues({'greeting': 'ciao'});
      final prefs = await SharedPreferences.getInstance();

      final container = ProviderContainer(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
      );
      addTearDown(container.dispose);

      final resolved = container.read(sharedPrefsProvider);
      expect(resolved.getString('greeting'), 'ciao');
    });

    test('override with data — stored bool value is accessible', () async {
      SharedPreferences.setMockInitialValues({'onboarding_completed': true});
      final prefs = await SharedPreferences.getInstance();

      final container = ProviderContainer(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
      );
      addTearDown(container.dispose);

      final resolved = container.read(sharedPrefsProvider);
      expect(resolved.getBool('onboarding_completed'), isTrue);
    });

    test('override with data — stored int value is accessible', () async {
      SharedPreferences.setMockInitialValues({'retry_count': 3});
      final prefs = await SharedPreferences.getInstance();

      final container = ProviderContainer(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
      );
      addTearDown(container.dispose);

      final resolved = container.read(sharedPrefsProvider);
      expect(resolved.getInt('retry_count'), 3);
    });

    test('override with empty prefs — missing key returns null', () async {
      SharedPreferences.setMockInitialValues({});
      final prefs = await SharedPreferences.getInstance();

      final container = ProviderContainer(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
      );
      addTearDown(container.dispose);

      final resolved = container.read(sharedPrefsProvider);
      expect(resolved.getString('nonexistent_key'), isNull);
    });

    test('two containers with independent overrides do not share state',
        () async {
      SharedPreferences.setMockInitialValues({'value': 'alpha'});
      final prefs1 = await SharedPreferences.getInstance();

      SharedPreferences.setMockInitialValues({'value': 'beta'});
      final prefs2 = await SharedPreferences.getInstance();

      final c1 = ProviderContainer(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs1)],
      );
      final c2 = ProviderContainer(
        overrides: [sharedPrefsProvider.overrideWithValue(prefs2)],
      );
      addTearDown(c1.dispose);
      addTearDown(c2.dispose);

      // Both containers resolve correctly to their respective prefs instances.
      expect(c1.read(sharedPrefsProvider), same(prefs1));
      expect(c2.read(sharedPrefsProvider), same(prefs2));
    });
  });
}
