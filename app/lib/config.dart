/// Runtime configuration — override at build time with --dart-define.
///
/// Dev (default):     flutter run
/// Production Render: flutter build apk --release \
///                      --dart-define=BACKEND_URL=wss://physicscopilot-api.onrender.com/ws
class AppConfig {
  const AppConfig._();

  static const String backendUrl = String.fromEnvironment(
    'BACKEND_URL',
    defaultValue: 'ws://localhost:8080/ws',
  );
}
