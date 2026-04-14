// App-wide constants (API URLs, timeouts, etc.)
//
// Override at build time with:
//   --dart-define=SERVER_URL=physicscopilot.onrender.com
//
// If SERVER_URL is not provided, falls back to the Cloudflare dev tunnel.
// Update [_devTunnelHost] whenever the tunnel is restarted, or override at
// runtime via [AppSettings.serverUrlOverride].

/// App-wide compile-time constants for server endpoints.
///
/// Pass the **hostname only** (no protocol, no trailing slash) via
/// `--dart-define=SERVER_URL=<host>`. The class prepends `https://` and
/// `wss://` automatically. Runtime overrides (Settings screen) still take
/// precedence via [AppSettings.serverUrlOverride].
///
/// Use [resolveApiUrl] and [resolveWsUrl] to build the correct URL from any
/// user-supplied string (accepts host-only, `https://`, or `wss://` formats).
class AppConstants {
  static const String _devTunnelHost =
      'tension-assume-portrait-pride.trycloudflare.com';

  /// Hostname injected at build time. Empty string means "use dev tunnel".
  static const String _serverHost = String.fromEnvironment(
    'SERVER_URL',
    defaultValue: '',
  );

  static const String _host =
      _serverHost == '' ? _devTunnelHost : _serverHost;

  /// Base WebSocket URL used when no runtime override is set.
  static const String wsBaseUrl = 'wss://$_host';

  /// Base HTTP URL used when no runtime override is set.
  static const String apiBaseUrl = 'https://$_host';

  // ---------------------------------------------------------------------------
  // URL resolvers — normalise user-supplied overrides
  // ---------------------------------------------------------------------------

  /// Returns the HTTPS base URL to use, honouring [override] if set.
  ///
  /// [override] may be:
  /// - `null` / empty → returns [apiBaseUrl]
  /// - `physicscopilot.onrender.com` → `https://physicscopilot.onrender.com`
  /// - `https://physicscopilot.onrender.com` → unchanged
  /// - `wss://physicscopilot.onrender.com` → converts to `https://`
  static String resolveApiUrl(String? override) {
    final host = _extractHost(override);
    return host != null ? 'https://$host' : apiBaseUrl;
  }

  /// Returns the WSS base URL to use, honouring [override] if set.
  ///
  /// [override] may be:
  /// - `null` / empty → returns [wsBaseUrl]
  /// - `physicscopilot.onrender.com` → `wss://physicscopilot.onrender.com`
  /// - `wss://physicscopilot.onrender.com` → unchanged
  /// - `https://physicscopilot.onrender.com` → converts to `wss://`
  static String resolveWsUrl(String? override) {
    final host = _extractHost(override);
    return host != null ? 'wss://$host' : wsBaseUrl;
  }

  /// Strips any `scheme://` prefix and trailing slash, returning the bare host.
  /// Returns `null` when [raw] is null or empty after trimming.
  static String? _extractHost(String? raw) {
    if (raw == null || raw.trim().isEmpty) return null;
    var s = raw.trim();
    final schemeEnd = s.indexOf('://');
    if (schemeEnd != -1) s = s.substring(schemeEnd + 3);
    if (s.endsWith('/')) s = s.substring(0, s.length - 1);
    return s.isEmpty ? null : s;
  }
}
