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
  group('SettingsNotifier', () {
    setUp(() {
      SharedPreferences.setMockInitialValues({});
    });

    test('initial state: voiceEnabled=true, themeMode=dark, serverUrlOverride=null', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      final state = container.read(settingsProvider);
      expect(state.voiceEnabled, isTrue);
      expect(state.themeMode, equals(ThemeMode.dark));
      expect(state.serverUrlOverride, isNull);
    });

    test('setVoiceEnabled(false) → voiceEnabled=false, persisted in prefs', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      await container.read(settingsProvider.notifier).setVoiceEnabled(false);

      final state = container.read(settingsProvider);
      expect(state.voiceEnabled, isFalse);
      expect(prefs.getBool('voice_enabled'), isFalse);
    });

    test('setThemeMode(light) → themeMode=light, persisted in prefs', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      await container.read(settingsProvider.notifier).setThemeMode(ThemeMode.light);

      final state = container.read(settingsProvider);
      expect(state.themeMode, equals(ThemeMode.light));
      expect(prefs.getString('theme_mode'), equals('light'));
    });

    test('setServerUrl("wss://example.com") → serverUrlOverride set and persisted', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      await container.read(settingsProvider.notifier).setServerUrl('wss://example.com');

      final state = container.read(settingsProvider);
      expect(state.serverUrlOverride, equals('wss://example.com'));
      expect(prefs.getString('server_url_override'), equals('wss://example.com'));
    });

    test('setServerUrl(null) → serverUrlOverride=null, key removed from prefs', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      // First set a URL, then clear it
      await container.read(settingsProvider.notifier).setServerUrl('wss://example.com');
      await container.read(settingsProvider.notifier).setServerUrl(null);

      final state = container.read(settingsProvider);
      expect(state.serverUrlOverride, isNull);
      expect(prefs.containsKey('server_url_override'), isFalse);
    });

    test('setServerUrl("") → same as null (empty string cleared)', () async {
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      // First set a URL
      await container.read(settingsProvider.notifier).setServerUrl('wss://example.com');
      // Clear with empty string
      await container.read(settingsProvider.notifier).setServerUrl('');

      final state = container.read(settingsProvider);
      expect(state.serverUrlOverride, isNull);
      expect(prefs.containsKey('server_url_override'), isFalse);
    });

    test('load from saved prefs: theme_mode=light → initial ThemeMode.light', () async {
      SharedPreferences.setMockInitialValues({'theme_mode': 'light'});
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      final state = container.read(settingsProvider);
      expect(state.themeMode, equals(ThemeMode.light));
    });

    test('load from saved prefs: voice_enabled=false → initial voiceEnabled=false', () async {
      SharedPreferences.setMockInitialValues({'voice_enabled': false});
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      final state = container.read(settingsProvider);
      expect(state.voiceEnabled, isFalse);
    });

    test('load from saved prefs: server_url_override set → initial serverUrlOverride non-null', () async {
      SharedPreferences.setMockInitialValues({'server_url_override': 'wss://saved.example.com'});
      final prefs = await SharedPreferences.getInstance();
      final container = _makeContainer(prefs);
      addTearDown(container.dispose);

      final state = container.read(settingsProvider);
      expect(state.serverUrlOverride, equals('wss://saved.example.com'));
    });
  });
}
