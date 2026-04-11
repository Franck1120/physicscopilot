// App-wide constants (API URLs, timeouts, etc.)
// For local WiFi testing, set to the server's LAN IP.
// Change this value before building the APK for device testing.
class AppConstants {
  static const String _serverHost = '192.168.0.198';
  static const int _serverPort = 8080;

  static const String wsBaseUrl = 'ws://$_serverHost:$_serverPort';
  static const String apiBaseUrl = 'http://$_serverHost:$_serverPort';
}
