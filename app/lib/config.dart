/// Runtime configuration — override at build time with `--dart-define`.
///
/// Dev (default):     `flutter run`
/// Production Render: `flutter build apk --release \`
///                      `--dart-define=BACKEND_URL=wss://physicscopilot-api.onrender.com/ws`
///
/// Prefer [AppConstants] for the Cloudflare-tunnel URL used during development.
/// [AppConfig] exists for the Render deployment where the URL is stable.
class AppConfig {
  const AppConfig._();

  /// WebSocket backend URL injected at build time.
  ///
  /// Defaults to `ws://localhost:8080/ws` for local development.
  static const String backendUrl = String.fromEnvironment(
    'BACKEND_URL',
    defaultValue: 'ws://localhost:8080/ws',
  );
}
