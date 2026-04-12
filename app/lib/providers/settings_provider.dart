import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'prefs_provider.dart';

const _kServerUrlKey = 'server_url_override';
const _kVoiceEnabledKey = 'voice_enabled';

/// User-configurable settings backed by SharedPreferences.
class AppSettings {
  /// Runtime override for the server URL.
  /// When null, [AppConstants.wsBaseUrl] / [AppConstants.apiBaseUrl] are used.
  final String? serverUrlOverride;

  /// Whether voice synthesis is active during sessions.
  final bool voiceEnabled;

  const AppSettings({
    this.serverUrlOverride,
    this.voiceEnabled = true,
  });

  AppSettings copyWith({
    String? Function()? serverUrlOverride,
    bool? voiceEnabled,
  }) =>
      AppSettings(
        serverUrlOverride: serverUrlOverride != null
            ? serverUrlOverride()
            : this.serverUrlOverride,
        voiceEnabled: voiceEnabled ?? this.voiceEnabled,
      );
}

class SettingsNotifier extends StateNotifier<AppSettings> {
  SettingsNotifier(this._prefs)
      : super(AppSettings(
          serverUrlOverride: _prefs.getString(_kServerUrlKey),
          voiceEnabled: _prefs.getBool(_kVoiceEnabledKey) ?? true,
        ));

  final SharedPreferences _prefs;

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

  Future<void> setVoiceEnabled(bool enabled) async {
    await _prefs.setBool(_kVoiceEnabledKey, enabled);
    state = state.copyWith(voiceEnabled: enabled);
  }
}

final settingsProvider =
    StateNotifierProvider<SettingsNotifier, AppSettings>((ref) {
  final prefs = ref.watch(sharedPrefsProvider);
  return SettingsNotifier(prefs);
});
