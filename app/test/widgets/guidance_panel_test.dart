// Widget tests for GuidancePanel.
//
// GuidancePanel is a ConsumerWidget that reads sessionProvider.
// We override it with a StateNotifierProvider backed by a pre-set SessionState
// so we can exercise each rendering branch without a real server.
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/providers/session_provider.dart';
import 'package:physicscopilot/widgets/guidance_panel.dart';

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/// Builds the widget tree under test with [sessionState] pre-loaded.
Widget _buildPanel({
  required SessionState sessionState,
  String? cachedResponse,
  bool isOffline = false,
}) {
  final controller = TextEditingController();
  return ProviderScope(
    overrides: [
      sessionProvider.overrideWith((ref) {
        final notifier = SessionNotifier();
        // Force the notifier into the desired state by replaying the right
        // method, or just return a notifier whose initial state we control.
        return notifier;
      }),
    ],
    child: MaterialApp(
      home: Scaffold(
        body: _FixedSessionScope(
          state: sessionState,
          child: GuidancePanel(
            textController: controller,
            onSendText: () {},
            cachedResponse: cachedResponse,
            isOffline: isOffline,
          ),
        ),
      ),
    ),
  );
}

/// Wraps the widget tree with a ProviderScope that holds a fixed SessionState.
class _FixedSessionScope extends StatelessWidget {
  const _FixedSessionScope({required this.state, required this.child});
  final SessionState state;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return ProviderScope(
      overrides: [
        sessionProvider.overrideWith((ref) {
          final notifier = _FixedSessionNotifier(state);
          return notifier;
        }),
      ],
      child: child,
    );
  }
}

class _FixedSessionNotifier extends SessionNotifier {
  _FixedSessionNotifier(SessionState s) {
    state = s;
  }
}

void main() {
  group('GuidancePanel', () {
    testWidgets('renders without crash in idle state', (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(),
      ),);
      await tester.pump();
      expect(find.byType(GuidancePanel), findsOneWidget);
    });

    testWidgets('shows idle placeholder text when state is empty', (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(),
      ),);
      await tester.pump();
      // AppStrings.sessionIdle text is rendered in the idle branch.
      // The exact string is defined in utils/strings.dart; we look for a
      // non-empty text widget inside the idle Center widget.
      expect(find.byKey(const ValueKey('idle')), findsOneWidget);
    });

    testWidgets('shows ThinkingIndicator when isProcessing is true',
        (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(isProcessing: true),
      ),);
      await tester.pump();
      expect(find.byType(ThinkingIndicator), findsOneWidget);
    });

    testWidgets('shows "L\'AI sta analizzando" label when processing',
        (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(isProcessing: true),
      ),);
      await tester.pump();
      expect(find.textContaining("L'AI sta analizzando"), findsWidgets);
    });

    testWidgets('shows responseText when session has a plain text response',
        (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(responseText: 'Hello from AI'),
      ),);
      // _TypewriterResponse animates char-by-char; let all timers fire.
      await tester.pumpAndSettle();
      expect(find.textContaining('Hello from AI'), findsOneWidget);
    });

    testWidgets('shows error icon and text when errorText is set',
        (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(errorText: 'Connection failed'),
      ),);
      await tester.pump();
      expect(find.byKey(const ValueKey('error')), findsOneWidget);
      expect(find.textContaining('Connection failed'), findsOneWidget);
    });

    testWidgets('shows offline banner when isOffline is true with cachedResponse',
        (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(),
        isOffline: true,
        cachedResponse: 'Cached AI response',
      ),);
      await tester.pump();
      expect(find.byKey(const ValueKey('offline')), findsOneWidget);
      expect(find.textContaining('offline'), findsWidgets);
      expect(find.textContaining('Cached AI response'), findsOneWidget);
    });

    testWidgets('shows cached response text in offline mode', (tester) async {
      const cached = 'Last known answer from the AI';
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(),
        isOffline: true,
        cachedResponse: cached,
      ),);
      await tester.pump();
      expect(find.text(cached), findsOneWidget);
    });

    testWidgets('isOffline without cachedResponse shows idle state',
        (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(),
        isOffline: true,
        cachedResponse: null,
      ),);
      await tester.pump();
      // No cached response → falls through to idle branch.
      expect(find.byKey(const ValueKey('idle')), findsOneWidget);
    });

    testWidgets('shows text input field', (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(),
      ),);
      await tester.pump();
      expect(find.byType(TextField), findsOneWidget);
    });

    testWidgets('shows send button', (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(),
      ),);
      await tester.pump();
      expect(find.byIcon(Icons.send_rounded), findsOneWidget);
    });

    testWidgets('shows streaming text widget when isStreaming is true',
        (tester) async {
      await tester.pumpWidget(_buildPanel(
        sessionState: const SessionState(
          isStreaming: true,
          streamingText: 'partial chunk',
        ),
      ),);
      await tester.pump();
      expect(find.byKey(const ValueKey('streaming')), findsOneWidget);
    });
  });
}
