import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/widgets/camera_preview_widget.dart';
import 'package:physicscopilot/providers/camera_provider.dart';

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

void main() {
  group('CameraPreviewWidget', () {
    // ── Loading state ─────────────────────────────────────────────────────

    testWidgets('loading state → shows CircularProgressIndicator',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            // Never completes → widget stays in loading state.
            cameraInitProvider.overrideWith(
              (ref) => Future.delayed(const Duration(hours: 1)),
            ),
          ],
          child: const MaterialApp(
            home: Scaffold(body: CameraPreviewWidget()),
          ),
        ),
      );

      // Single pump: FutureProvider is still pending.
      await tester.pump();

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    // ── Error state ───────────────────────────────────────────────────────

    testWidgets('error state → shows "Camera non disponibile" text',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            cameraInitProvider.overrideWith(
              (ref) => Future.error('camera error'),
            ),
          ],
          child: const MaterialApp(
            home: Scaffold(body: CameraPreviewWidget()),
          ),
        ),
      );

      // First pump: FutureProvider is pending (loading).
      await tester.pump();
      // Second pump: Future.error resolves → error branch rendered.
      await tester.pump();

      expect(find.text('Camera non disponibile'), findsOneWidget);
    });

    testWidgets('error state → does NOT show CircularProgressIndicator',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            cameraInitProvider.overrideWith(
              (ref) => Future.error('camera error'),
            ),
          ],
          child: const MaterialApp(
            home: Scaffold(body: CameraPreviewWidget()),
          ),
        ),
      );

      await tester.pump();
      await tester.pump();

      expect(find.byType(CircularProgressIndicator), findsNothing);
    });

    testWidgets('error state → shows camera_alt icon', (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            cameraInitProvider.overrideWith(
              (ref) => Future.error('camera error'),
            ),
          ],
          child: const MaterialApp(
            home: Scaffold(body: CameraPreviewWidget()),
          ),
        ),
      );

      await tester.pump();
      await tester.pump();

      expect(find.byIcon(Icons.camera_alt), findsOneWidget);
    });

    // ── Loading → error transition ────────────────────────────────────────

    testWidgets(
        'starts in loading state then transitions to error state on Future.error',
        (tester) async {
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            cameraInitProvider.overrideWith(
              (ref) => Future.error('init failed'),
            ),
          ],
          child: const MaterialApp(
            home: Scaffold(body: CameraPreviewWidget()),
          ),
        ),
      );

      // Before the Future resolves: loading spinner visible.
      await tester.pump();
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Camera non disponibile'), findsNothing);

      // After the Future resolves with an error: error UI visible.
      await tester.pump();
      expect(find.byType(CircularProgressIndicator), findsNothing);
      expect(find.text('Camera non disponibile'), findsOneWidget);
    });
  });
}
