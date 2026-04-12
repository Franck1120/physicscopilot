// Widget tests for LoginScreen.
//
// These tests verify form validation and the sign-in / sign-up toggle.
// No Supabase calls are made: validation errors are caught before AuthService
// is invoked, and the mock values ensure SharedPreferences is not needed.
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/screens/login_screen.dart';

void main() {
  Widget buildLoginScreen() => const MaterialApp(home: LoginScreen());

  group('LoginScreen — initial render', () {
    testWidgets('shows sign-in mode title and subtitle by default', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      // Title (rendered as Text inside the Column)
      expect(find.text('Accedi'), findsWidgets);
      expect(find.text('Bentornato su PhysicsCopilot'), findsOneWidget);
    });

    testWidgets('shows toggle hint to switch to sign-up', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      expect(find.text('Non hai un account? Registrati'), findsOneWidget);
    });

    testWidgets('password field is obscured by default', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      expect(find.byIcon(Icons.visibility_off_outlined), findsOneWidget);
    });
  });

  group('LoginScreen — sign-in / sign-up toggle', () {
    testWidgets('tapping toggle switches to sign-up mode', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      await tester.tap(find.text('Non hai un account? Registrati'));
      await tester.pump();

      expect(find.text('Crea account'), findsOneWidget);
      expect(find.text('Inizia ad usare PhysicsCopilot'), findsOneWidget);
      expect(find.text('Hai gia un account? Accedi'), findsOneWidget);
    });

    testWidgets('tapping toggle twice returns to sign-in mode', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      // → sign-up
      await tester.tap(find.text('Non hai un account? Registrati'));
      await tester.pump();

      // → sign-in
      await tester.tap(find.text('Hai gia un account? Accedi'));
      await tester.pump();

      expect(find.text('Accedi'), findsWidgets);
      expect(find.text('Bentornato su PhysicsCopilot'), findsOneWidget);
    });

    testWidgets('sign-up submit button shows "Registrati"', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      await tester.tap(find.text('Non hai un account? Registrati'));
      await tester.pump();

      // ElevatedButton should now show "Registrati".
      expect(find.widgetWithText(ElevatedButton, 'Registrati'), findsOneWidget);
    });
  });

  group('LoginScreen — form validation', () {
    testWidgets('shows error when email is empty on submit', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      await tester.tap(find.widgetWithText(ElevatedButton, 'Accedi'));
      await tester.pump();

      expect(find.text('Inserisci la tua email'), findsOneWidget);
    });

    testWidgets('shows error when email does not contain @', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      // Email field is the first TextFormField.
      await tester.enterText(
        find.byType(TextFormField).first,
        'notemail.invalid',
      );
      await tester.tap(find.widgetWithText(ElevatedButton, 'Accedi'));
      await tester.pump();

      expect(find.text('Email non valida'), findsOneWidget);
    });

    testWidgets('shows error when password is empty on submit', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      await tester.enterText(
        find.byType(TextFormField).first,
        'test@example.com',
      );
      await tester.tap(find.widgetWithText(ElevatedButton, 'Accedi'));
      await tester.pump();

      expect(find.text('Inserisci la tua password'), findsOneWidget);
    });

    testWidgets('shows min-length error for password shorter than 8 chars in sign-up', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      // Switch to sign-up mode.
      await tester.tap(find.text('Non hai un account? Registrati'));
      await tester.pump();

      await tester.enterText(
        find.byType(TextFormField).first,
        'user@example.com',
      );
      // Password field is the last TextFormField.
      await tester.enterText(
        find.byType(TextFormField).last,
        'short',
      );
      await tester.tap(find.widgetWithText(ElevatedButton, 'Registrati'));
      await tester.pump();

      expect(find.text('Minimo 8 caratteri'), findsOneWidget);
    });

    testWidgets('no validation error for valid email and password in sign-in mode', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      await tester.enterText(
        find.byType(TextFormField).first,
        'valid@example.com',
      );
      await tester.enterText(
        find.byType(TextFormField).last,
        'validpass',
      );
      await tester.tap(find.widgetWithText(ElevatedButton, 'Accedi'));
      await tester.pump();

      // No validation error texts should appear (submit proceeds to AuthService,
      // which throws because Supabase is not configured; the error is swallowed
      // and the loading indicator disappears, but no validation text is shown).
      expect(find.text('Inserisci la tua email'), findsNothing);
      expect(find.text('Email non valida'), findsNothing);
      expect(find.text('Inserisci la tua password'), findsNothing);
    });
  });

  group('LoginScreen — password visibility toggle', () {
    testWidgets('tapping visibility icon reveals password', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      // Initially obscured — off icon visible.
      expect(find.byIcon(Icons.visibility_off_outlined), findsOneWidget);

      await tester.tap(find.byIcon(Icons.visibility_off_outlined));
      await tester.pump();

      // Now revealed — on icon visible, off icon gone.
      expect(find.byIcon(Icons.visibility_outlined), findsOneWidget);
      expect(find.byIcon(Icons.visibility_off_outlined), findsNothing);
    });

    testWidgets('tapping visibility icon a second time re-obscures password', (tester) async {
      await tester.pumpWidget(buildLoginScreen());
      await tester.pump();

      await tester.tap(find.byIcon(Icons.visibility_off_outlined));
      await tester.pump();

      await tester.tap(find.byIcon(Icons.visibility_outlined));
      await tester.pump();

      expect(find.byIcon(Icons.visibility_off_outlined), findsOneWidget);
    });
  });
}
