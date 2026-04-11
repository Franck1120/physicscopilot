// App-wide constants (API URLs, timeouts, etc.)
// Public tunnel via Cloudflare (no port, HTTPS/WSS).
// Update this URL whenever the tunnel is restarted.
class AppConstants {
  static const String _tunnelHost = 'tension-assume-portrait-pride.trycloudflare.com';

  static const String wsBaseUrl = 'wss://$_tunnelHost';
  static const String apiBaseUrl = 'https://$_tunnelHost';
}
