// Tests for app theme state via SettingsProvider.
//
// There is no dedicated themeProvider — theme mode is stored inside
// SettingsNotifier (settingsProvider). These tests verify that the
// SharedPreferences round-trip for ThemeMode works correctly and that
// the prefs provider delivers the expected SharedPreferences instance
// to downstream providers.

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:physicscopilot/providers/prefs_provider.dart';
import 'package:physicscopilot/providers/settings_provider.dart';

ProviderContainer _makeContainer(SharedPreferences prefs) {
  return ProviderContainer(
    overrides: [sharedPrefsProvider.overrideWithValue(prefs)],
  );
}

void main() {
  group('Theme/Prefs integration', () {
    setUp(() {
      SharedPreferences.setMockInitialValues({});
    });

    test('prefs provider returns the overridden SharedPreferences instance',
        () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      final storedPrefs = container.read(sharedPrefsProvider);
      expect(storedPrefs, same(prefs));
    });

    test('default themeMode from settingsProvider is ThemeMode.dark', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      final settings = container.read(settingsProvider);
      expect(settings.themeMode, equals(ThemeMode.dark));
    });

    test('setThemeMode(light) persists and is immediately readable', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      await container.read(settingsProvider.notifier).setThemeMode(ThemeMode.light);

      expect(container.read(settingsProvider).themeMode, equals(ThemeMode.light));
      expect(prefs.getString('theme_mode'), equals('light'));
    });

    test('setThemeMode(system) persists and is immediately readable', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      await container.read(settingsProvider.notifier).setThemeMode(ThemeMode.system);

      expect(container.read(settingsProvider).themeMode, equals(ThemeMode.system));
      expect(prefs.getString('theme_mode'), equals('system'));
    });

    test('restoring ThemeMode.dark: dark → light → dark round-trip', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      await container.read(settingsProvider.notifier).setThemeMode(ThemeMode.light);
      await container.read(settingsProvider.notifier).setThemeMode(ThemeMode.dark);

      expect(container.read(settingsProvider).themeMode, equals(ThemeMode.dark));
      expect(prefs.getString('theme_mode'), equals('dark'));
    });

    test('theme_mode=light in saved prefs → initial ThemeMode.light on cold start',
        () async {
      SharedPreferences.setMockInitialValues({'theme_mode': 'light'});
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      expect(container.read(settingsProvider).themeMode, equals(ThemeMode.light));
    });

    test('theme_mode=system in saved prefs → initial ThemeMode.system on cold start',
        () async {
      SharedPreferences.setMockInitialValues({'theme_mode': 'system'});
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      expect(container.read(settingsProvider).themeMode, equals(ThemeMode.system));
    });
  });
}
