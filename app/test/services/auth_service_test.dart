// Unit tests for AuthService.
//
// AuthService uses only static methods backed by Supabase.instance.client.
// In test, Supabase is never initialised (no --dart-define env vars), so
// _client returns null. We verify the null-safety behaviour (guard clauses,
// thrown exceptions, safe no-ops) and the initialize() early-return path.
import 'package:flutter_test/flutter_test.dart';
import 'package:supabase_flutter/supabase_flutter.dart';

import 'package:physicscopilot/services/auth_service.dart';

void main() {
  group('AuthService — when Supabase is not initialised', () {
    // ── currentUser / isAuthenticated ──────────────────────────────────────

    test('currentUser returns null when Supabase is not initialised', () {
      expect(AuthService.currentUser, isNull);
    });

    test('isAuthenticated returns false when Supabase is not initialised', () {
      expect(AuthService.isAuthenticated, isFalse);
    });

    // ── Access token ──────────────────────────────────────────────────────

    test('currentAccessToken returns null when Supabase is not initialised',
        () {
      expect(AuthService.currentAccessToken, isNull);
    });

    test('getAccessToken() returns null when Supabase is not initialised',
        () async {
      final token = await AuthService.getAccessToken();
      expect(token, isNull);
    });

    // ── authStateChanges ──────────────────────────────────────────────────

    test(
        'authStateChanges emits nothing when Supabase is not initialised',
        () async {
      // The fallback is Stream.empty(), which completes immediately.
      final events = await AuthService.authStateChanges.toList();
      expect(events, isEmpty);
    });

    // ── signInWithEmail ───────────────────────────────────────────────────

    test(
        'signInWithEmail throws AuthException when Supabase is not initialised',
        () async {
      expect(
        () => AuthService.signInWithEmail('test@example.com', 'password123'),
        throwsA(isA<AuthException>().having(
          (e) => e.message,
          'message',
          contains('not initialised'),
        ),),
      );
    });

    // ── signUpWithEmail ───────────────────────────────────────────────────

    test(
        'signUpWithEmail throws AuthException when Supabase is not initialised',
        () async {
      expect(
        () => AuthService.signUpWithEmail('test@example.com', 'password123'),
        throwsA(isA<AuthException>().having(
          (e) => e.message,
          'message',
          contains('not initialised'),
        ),),
      );
    });

    // ── signOut ───────────────────────────────────────────────────────────

    test('signOut completes without error when Supabase is not initialised',
        () async {
      // signOut uses _client?.auth.signOut() — null-aware, so it's a no-op.
      await expectLater(AuthService.signOut(), completes);
    });
  });

  group('AuthService — initialize()', () {
    test(
        'initialize() returns immediately when env vars are empty '
        '(no Supabase client created)', () async {
      // SUPABASE_URL and SUPABASE_ANON_KEY are not set in test, so
      // initialize() should early-return without calling Supabase.initialize.
      await expectLater(AuthService.initialize(), completes);

      // After early-return, Supabase is still not initialised.
      expect(AuthService.isAuthenticated, isFalse);
      expect(AuthService.currentUser, isNull);
    });
  });
}
