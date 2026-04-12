import 'package:flutter/material.dart' show ThemeMode;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'prefs_provider.dart';

const _kServerUrlKey = 'server_url_override';
const _kVoiceEnabledKey = 'voice_enabled';
const _kLanguageKey = 'language';
const _kThemeModeKey = 'theme_mode'; // 'dark' | 'light'

/// Supported response languages (BCP-47 code to display name).
const Map<String, String> kSupportedLanguages = {
  'it': 'Italiano',
  'en': 'English',
  'fr': 'Francais',
  'de': 'Deutsch',
  'es': 'Espanol',
};

/// User-configurable settings backed by SharedPreferences.
class AppSettings {
  final String? serverUrlOverride;
  final bool voiceEnabled;
  /// BCP-47 language code for Gemini responses and TTS (default: "it").
  final String language;
  /// App-wide theme mode (dark by default).
  final ThemeMode themeMode;

  const AppSettings({
    this.serverUrlOverride,
    this.voiceEnabled = true,
    this.language = 'it',
    this.themeMode = ThemeMode.dark,
  });

  AppSettings copyWith({
    String? Function()? serverUrlOverride,
    bool? voiceEnabled,
    String? language,
    ThemeMode? themeMode,
  }) =>
      AppSettings(
        serverUrlOverride: serverUrlOverride != null
            ? serverUrlOverride()
            : this.serverUrlOverride,
        voiceEnabled: voiceEnabled ?? this.voiceEnabled,
        language: language ?? this.language,
        themeMode: themeMode ?? this.themeMode,
      );
}

/// Manages [AppSettings] and persists changes to SharedPreferences.
class SettingsNotifier extends StateNotifier<AppSettings> {
  SettingsNotifier(this._prefs)
      : super(AppSettings(
          serverUrlOverride: _prefs.getString(_kServerUrlKey),
          voiceEnabled: _prefs.getBool(_kVoiceEnabledKey) ?? true,
          language: _prefs.getString(_kLanguageKey) ?? 'it',
          themeMode: _prefs.getString(_kThemeModeKey) == 'light'
              ? ThemeMode.light
              : ThemeMode.dark,
        ));

  final SharedPreferences _prefs;

  /// Persists the server URL override. Pass `null` or empty string to reset.
  Future<void> setServerUrl(String? url) async {
    final trimmed = url?.trim();
    if (trimmed == null || trimmed.isEmpty) {
      await _prefs.remove(_kServerUrlKey);
      state = state.copyWith(serverUrlOverride: () => null);
    } else {
      await _prefs.setString(_kServerUrlKey, trimmed);
      state = state.copyWith(serverUrlOverride: () => trimmed);
    }
  }

  /// Persists the voice-guidance toggle.
  Future<void> setVoiceEnabled(bool enabled) async {
    await _prefs.setBool(_kVoiceEnabledKey, enabled);
    state = state.copyWith(voiceEnabled: enabled);
  }

  /// Persists the BCP-47 [lang] code for AI responses and TTS.
  Future<void> setLanguage(String lang) async {
    await _prefs.setString(_kLanguageKey, lang);
    state = state.copyWith(language: lang);
  }

  /// Persists the app-wide theme mode.
  Future<void> setThemeMode(ThemeMode mode) async {
    await _prefs.setString(
        _kThemeModeKey, mode == ThemeMode.light ? 'light' : 'dark');
    state = state.copyWith(themeMode: mode);
  }
}

final settingsProvider =
    StateNotifierProvider<SettingsNotifier, AppSettings>((ref) {
  final prefs = ref.watch(sharedPrefsProvider);
  return SettingsNotifier(prefs);
});
