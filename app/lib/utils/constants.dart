// App-wide constants (API URLs, timeouts, etc.)
//
// SERVER_URL is injected at build time via --dart-define:
//
//   # Local development (plain WS/HTTP):
//   flutter run --dart-define=SERVER_URL=localhost:8080
//
//   # Production (HTTPS/WSS inferred automatically):
//   flutter build apk --dart-define=SERVER_URL=your.domain.com
//
// If SERVER_URL is not provided the app defaults to localhost:8080 so that
// `flutter run` without arguments works out of the box.
class AppConstants {
  // Compile-time constant — set via --dart-define=SERVER_URL=<host:port>
  static const String _serverHost = String.fromEnvironment(
    'SERVER_URL',
    defaultValue: 'localhost:8080',
  );

  // localhost / 127.x addresses use plain WS/HTTP; everything else uses WSS/HTTPS.
  static bool get _secure =>
      !_serverHost.startsWith('localhost') &&
      !_serverHost.startsWith('127.');

  static String get wsBaseUrl =>
      _secure ? 'wss://$_serverHost' : 'ws://$_serverHost';

  static String get apiBaseUrl =>
      _secure ? 'https://$_serverHost' : 'http://$_serverHost';
}
