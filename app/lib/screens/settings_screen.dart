// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../main.dart'
    show kAccent, kBgPrimary, kBgCard, kBgCardBorder, kTextMuted;
import '../widgets/safe_screen.dart';
import '../providers/settings_provider.dart';
import '../providers/voice_provider.dart';
import '../utils/constants.dart';

// ThemeMode is imported via flutter/material.dart above.

/// App settings screen.
///
/// Accessible from the Profile tab (Impostazioni tile).
/// Allows the user to:
/// - Override the server URL at runtime (persisted in SharedPreferences)
/// - Toggle voice synthesis on/off
/// - View app version and build info
class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  late final TextEditingController _urlController;
  bool _urlEdited = false;

  @override
  void initState() {
    super.initState();
    final current = ref.read(settingsProvider).serverUrlOverride;
    _urlController = TextEditingController(text: current ?? '');
  }

  @override
  void dispose() {
    _urlController.dispose();
    super.dispose();
  }

  Future<void> _saveUrl() async {
    HapticFeedback.lightImpact();
    await ref
        .read(settingsProvider.notifier)
        .setServerUrl(_urlController.text);
    setState(() => _urlEdited = false);
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('URL server aggiornato — riavvia la sessione.'),
          behavior: SnackBarBehavior.floating,
        ),
      );
    }
  }

  Future<void> _resetUrl() async {
    HapticFeedback.lightImpact();
    _urlController.clear();
    await ref.read(settingsProvider.notifier).setServerUrl(null);
    setState(() => _urlEdited = false);
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('URL ripristinato al valore di default.'),
          behavior: SnackBarBehavior.floating,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    try {
      return _buildContent(context);
    } catch (e) {
      return screenError(e, context);
    }
  }

  Widget _buildContent(BuildContext context) {
    final settings = ref.watch(settingsProvider);

    return Scaffold(
      backgroundColor: kBgPrimary,
      appBar: AppBar(
        backgroundColor: const Color(0xFF111111),
        elevation: 0,
        title: const Text(
          'Impostazioni',
          style: TextStyle(
              color: Colors.white,
              fontWeight: FontWeight.bold,
              letterSpacing: 0.4),
        ),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new,
              color: Colors.white, size: 20),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: ListView(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
        children: [
          // ── Server URL ────────────────────────────────────────────────────
          _SectionHeader(label: 'CONNESSIONE'),
          const SizedBox(height: 10),
          _Card(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'URL Server',
                  style: TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      fontWeight: FontWeight.w600),
                ),
                const SizedBox(height: 4),
                Text(
                  'Default: ${AppConstants.wsBaseUrl}\nLascia vuoto per usare il valore compilato.',
                  style: const TextStyle(color: kTextMuted, fontSize: 12, height: 1.4),
                ),
                const SizedBox(height: 12),
                Semantics(
                  label: 'URL del server, lascia vuoto per usare il default',
                  child: TextField(
                    controller: _urlController,
                    style:
                        const TextStyle(color: Colors.white, fontSize: 13),
                    keyboardType: TextInputType.url,
                    autocorrect: false,
                    onChanged: (_) => setState(() => _urlEdited = true),
                    decoration: InputDecoration(
                      hintText: 'es. wss://your-tunnel.trycloudflare.com',
                      hintStyle:
                          const TextStyle(color: kTextMuted, fontSize: 12),
                      filled: true,
                      fillColor: const Color(0xFF111111),
                      contentPadding: const EdgeInsets.symmetric(
                          horizontal: 14, vertical: 10),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(10),
                        borderSide:
                            const BorderSide(color: kBgCardBorder),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(10),
                        borderSide:
                            const BorderSide(color: kBgCardBorder),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(10),
                        borderSide:
                            const BorderSide(color: kAccent),
                      ),
                      suffixIcon: _urlController.text.isNotEmpty
                          ? IconButton(
                              icon: const Icon(Icons.clear,
                                  color: kTextMuted, size: 18),
                              onPressed: () {
                                _urlController.clear();
                                setState(() => _urlEdited = true);
                              },
                            )
                          : null,
                    ),
                  ),
                ),
                const SizedBox(height: 10),
                Row(
                  children: [
                    if (settings.serverUrlOverride != null)
                      Expanded(
                        child: OutlinedButton(
                          onPressed: _resetUrl,
                          style: OutlinedButton.styleFrom(
                            foregroundColor: Colors.redAccent,
                            side: const BorderSide(color: Colors.redAccent),
                          ),
                          child: const Text('Reset'),
                        ),
                      ),
                    if (settings.serverUrlOverride != null)
                      const SizedBox(width: 10),
                    Expanded(
                      child: ElevatedButton(
                        onPressed: _urlEdited ? _saveUrl : null,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: kAccent,
                          foregroundColor: Colors.white,
                          disabledBackgroundColor: kAccent.withAlpha(40),
                        ),
                        child: const Text('Salva'),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          const SizedBox(height: 24),

          // ── Voice + Theme ─────────────────────────────────────────────────
          _SectionHeader(label: 'FUNZIONALITÀ'),
          const SizedBox(height: 10),
          _Card(
            child: Column(
              children: [
                // Voice toggle
                Row(
                  children: [
                    const Icon(Icons.volume_up_outlined,
                        color: kAccent, size: 22),
                    const SizedBox(width: 14),
                    const Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Guida vocale',
                              style: TextStyle(
                                  color: Colors.white,
                                  fontSize: 14,
                                  fontWeight: FontWeight.w600)),
                          SizedBox(height: 2),
                          Text('Legge le istruzioni AI ad alta voce.',
                              style:
                                  TextStyle(color: kTextMuted, fontSize: 12)),
                        ],
                      ),
                    ),
                    Semantics(
                      label:
                          'Guida vocale ${settings.voiceEnabled ? "attiva" : "disattiva"}',
                      toggled: settings.voiceEnabled,
                      child: Switch(
                        value: settings.voiceEnabled,
                        onChanged: (v) => ref
                            .read(settingsProvider.notifier)
                            .setVoiceEnabled(v),
                        activeThumbColor: kAccent,
                      ),
                    ),
                  ],
                ),
                const Divider(color: kBgCardBorder, height: 20),
                // Theme toggle
                Row(
                  children: [
                    Icon(
                      settings.themeMode == ThemeMode.dark
                          ? Icons.dark_mode_outlined
                          : Icons.light_mode_outlined,
                      color: kAccent,
                      size: 22,
                    ),
                    const SizedBox(width: 14),
                    const Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Tema scuro',
                              style: TextStyle(
                                  color: Colors.white,
                                  fontSize: 14,
                                  fontWeight: FontWeight.w600)),
                          SizedBox(height: 2),
                          Text('Passa tra tema chiaro e scuro.',
                              style:
                                  TextStyle(color: kTextMuted, fontSize: 12)),
                        ],
                      ),
                    ),
                    Semantics(
                      label:
                          'Tema ${settings.themeMode == ThemeMode.dark ? "scuro attivo" : "chiaro attivo"}',
                      toggled: settings.themeMode == ThemeMode.dark,
                      child: Switch(
                        value: settings.themeMode == ThemeMode.dark,
                        onChanged: (v) => ref
                            .read(settingsProvider.notifier)
                            .setThemeMode(
                                v ? ThemeMode.dark : ThemeMode.light),
                        activeThumbColor: kAccent,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          const SizedBox(height: 16),

          // ── Language ──────────────────────────────────────────────────────
          _Card(
            child: Row(
              children: [
                const Icon(Icons.language_outlined, color: kAccent, size: 22),
                const SizedBox(width: 14),
                const Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Lingua risposta AI',
                          style: TextStyle(
                              color: Colors.white,
                              fontSize: 14,
                              fontWeight: FontWeight.w600)),
                      SizedBox(height: 2),
                      Text('Lingua di Gemini e guida vocale.',
                          style: TextStyle(color: kTextMuted, fontSize: 12)),
                    ],
                  ),
                ),
                DropdownButton<String>(
                  value: settings.language,
                  dropdownColor: kBgCard,
                  underline: const SizedBox.shrink(),
                  icon: const Icon(Icons.expand_more, color: kTextMuted, size: 18),
                  style: const TextStyle(color: Colors.white, fontSize: 13),
                  items: kSupportedLanguages.entries
                      .map((e) => DropdownMenuItem(
                            value: e.key,
                            child: Text(e.value),
                          ))
                      .toList(),
                  onChanged: (lang) async {
                    if (lang == null) return;
                    HapticFeedback.selectionClick();
                    await ref.read(settingsProvider.notifier).setLanguage(lang);
                    // Update TTS language immediately.
                    ref.read(voiceServiceProvider).setLanguage(lang);
                  },
                ),
              ],
            ),
          ),

          const SizedBox(height: 24),

          // ── Info versione ─────────────────────────────────────────────────
          _SectionHeader(label: 'CONNESSIONE SERVER'),
          const SizedBox(height: 10),
          _Card(
            child: Column(
              children: [
                _InfoRow(
                  label: 'URL compilato',
                  value: AppConstants.wsBaseUrl,
                ),
                if (settings.serverUrlOverride != null) ...[
                  const Divider(color: kBgCardBorder, height: 20),
                  _InfoRow(
                    label: 'Override attivo',
                    value: settings.serverUrlOverride!,
                    valueColor: kAccent,
                  ),
                ],
              ],
            ),
          ),

          const SizedBox(height: 24),

          // ── About ─────────────────────────────────────────────────────────
          _SectionHeader(label: 'INFORMAZIONI APP'),
          const SizedBox(height: 10),
          _Card(
            child: Column(
              children: [
                _InfoRow(label: 'App', value: 'PhysicsCopilot'),
                const Divider(color: kBgCardBorder, height: 20),
                _InfoRow(label: 'Versione', value: '1.0.0 (build 1)'),
                const Divider(color: kBgCardBorder, height: 20),
                _InfoRow(label: 'Motore AI', value: 'Google Gemini'),
                const Divider(color: kBgCardBorder, height: 20),
                Align(
                  alignment: Alignment.centerLeft,
                  child: TextButton.icon(
                    onPressed: () {
                      HapticFeedback.selectionClick();
                      showAboutAppDialog(context);
                    },
                    icon: const Icon(Icons.info_outline, size: 16),
                    label: const Text('Dettagli e licenze'),
                    style: TextButton.styleFrom(
                      foregroundColor: kAccent,
                      padding: EdgeInsets.zero,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// About dialog — shared between SettingsScreen and ProfileTab
// ---------------------------------------------------------------------------

/// Shows the app's About dialog with version, credits, and policy links.
void showAboutAppDialog(BuildContext context) {
  showDialog<void>(
    context: context,
    builder: (_) => AlertDialog(
      backgroundColor: kBgCard,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: const BorderSide(color: kBgCardBorder, width: 1),
      ),
      title: const Row(
        children: [
          Icon(Icons.science_rounded, color: kAccent, size: 22),
          SizedBox(width: 10),
          Text(
            'PhysicsCopilot',
            style: TextStyle(color: Colors.white, fontSize: 18),
          ),
        ],
      ),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Versione 1.0.0 (build 1)',
            style: TextStyle(color: kTextMuted, fontSize: 13),
          ),
          const SizedBox(height: 16),
          const Text(
            'Powered by',
            style: TextStyle(color: kTextMuted, fontSize: 11, letterSpacing: 0.4),
          ),
          const SizedBox(height: 4),
          const Text(
            'Google Gemini AI',
            style: TextStyle(
              color: Colors.white,
              fontSize: 14,
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(height: 20),
          const Divider(color: kBgCardBorder, height: 1),
          const SizedBox(height: 16),
          const Text(
            'Privacy Policy',
            style: TextStyle(
              color: kAccent,
              fontSize: 13,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 4),
          const SelectableText(
            'https://physicscopilot.app/privacy',
            style: TextStyle(color: kTextMuted, fontSize: 12),
          ),
          const SizedBox(height: 14),
          const Text(
            'Termini di Servizio',
            style: TextStyle(
              color: kAccent,
              fontSize: 13,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 4),
          const SelectableText(
            'https://physicscopilot.app/terms',
            style: TextStyle(color: kTextMuted, fontSize: 12),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Chiudi', style: TextStyle(color: kAccent)),
        ),
      ],
    ),
  );
}

// ── Shared sub-widgets ────────────────────────────────────────────────────────

class _SectionHeader extends StatelessWidget {
  const _SectionHeader({required this.label});
  final String label;

  @override
  Widget build(BuildContext context) => Text(
        label,
        style: const TextStyle(
          color: kTextMuted,
          fontSize: 11,
          fontWeight: FontWeight.w600,
          letterSpacing: 0.8,
        ),
      );
}

class _Card extends StatelessWidget {
  const _Card({required this.child});
  final Widget child;

  @override
  Widget build(BuildContext context) => Container(
        width: double.infinity,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: kBgCard,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: kBgCardBorder, width: 1),
        ),
        child: child,
      );
}

class _InfoRow extends StatelessWidget {
  const _InfoRow({
    required this.label,
    required this.value,
    this.valueColor,
  });
  final String label;
  final String value;
  final Color? valueColor;

  @override
  Widget build(BuildContext context) => Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 140,
            child: Text(label,
                style: const TextStyle(color: kTextMuted, fontSize: 13)),
          ),
          Expanded(
            child: Text(
              value,
              style: TextStyle(
                color: valueColor ?? Colors.white,
                fontSize: 13,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      );
}
