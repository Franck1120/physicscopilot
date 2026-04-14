// App-wide constants (API URLs, timeouts, etc.)
// Public tunnel via Cloudflare (no port, HTTPS/WSS).
// Update this URL whenever the tunnel is restarted.

/// App-wide compile-time constants for server endpoints.
///
/// The default URLs point to the active Cloudflare tunnel.
/// Update [_tunnelHost] whenever the tunnel is restarted, or override at
/// runtime via [AppSettings.serverUrlOverride].
class AppConstants {
  static const String _tunnelHost = 'physicscopilot.onrender.com';

  /// Base WebSocket URL used when no runtime override is set.
  static const String wsBaseUrl = 'wss://$_tunnelHost';

  /// Base HTTP URL used when no runtime override is set.
  static const String apiBaseUrl = 'https://$_tunnelHost';
}
