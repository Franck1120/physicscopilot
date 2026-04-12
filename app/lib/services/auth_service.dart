import 'package:supabase_flutter/supabase_flutter.dart';

/// Wraps Supabase auth operations used throughout the app.
///
/// URL and anon key are injected via --dart-define at build time:
///   flutter run \
///     --dart-define=SUPABASE_URL=https://xxxx.supabase.co \
///     --dart-define=SUPABASE_ANON_KEY=eyJ...
///
/// Call [AuthService.initialize] once before [runApp].
class AuthService {
  static const String _supabaseUrl = String.fromEnvironment('SUPABASE_URL');
  static const String _supabaseAnonKey =
      String.fromEnvironment('SUPABASE_ANON_KEY');

  /// Initialises the Supabase client. Must be called in [main] before [runApp].
  static Future<void> initialize() async {
    if (_supabaseUrl.isEmpty || _supabaseAnonKey.isEmpty) {
      // Dev mode: Supabase not configured — auth is skipped.
      return;
    }
    await Supabase.initialize(url: _supabaseUrl, anonKey: _supabaseAnonKey);
  }

  /// Returns the Supabase client, or null when not initialised.
  static SupabaseClient? get _client {
    try {
      return Supabase.instance.client;
    } catch (_) {
      return null;
    }
  }

  /// The currently signed-in user, or null when not authenticated / not initialised.
  static User? get currentUser => _client?.auth.currentUser;

  /// True when a user is signed in.
  static bool get isAuthenticated => currentUser != null;

  /// Returns the JWT access token synchronously (for use in sync Providers).
  static String? get currentAccessToken =>
      _client?.auth.currentSession?.accessToken;

  /// Returns the JWT access token of the current session asynchronously.
  static Future<String?> getAccessToken() async =>
      _client?.auth.currentSession?.accessToken;

  /// Emits auth state changes (sign-in, sign-out, token refresh).
  /// Callers can use this to re-run GoRouter redirect guards.
  static Stream<AuthState> get authStateChanges =>
      _client?.auth.onAuthStateChange ?? const Stream.empty();

  /// Signs in with email + password.
  /// Throws [AuthException] on failure.
  static Future<AuthResponse> signInWithEmail(
      String email, String password) async {
    final client = _client;
    if (client == null) throw const AuthException('Supabase not initialised');
    return client.auth.signInWithPassword(email: email, password: password);
  }

  /// Creates a new account with email + password.
  /// Throws [AuthException] on failure.
  static Future<AuthResponse> signUpWithEmail(
      String email, String password) async {
    final client = _client;
    if (client == null) throw const AuthException('Supabase not initialised');
    return client.auth.signUp(email: email, password: password);
  }

  /// Signs out the current user.
  static Future<void> signOut() async {
    await _client?.auth.signOut();
  }
}
