import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'screens/onboarding_screen.dart';
import 'screens/printer_selection_screen.dart';
import 'screens/home_screen.dart';
import 'screens/camera_screen.dart';
import 'screens/history_screen.dart';
import 'providers/printer_provider.dart';

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

/// Exposes the SharedPreferences instance initialised before runApp.
final _sharedPrefsProvider = Provider<SharedPreferences>((ref) {
  throw UnimplementedError('Override with ProviderScope.overrides');
});

/// True if the user has already completed onboarding.
final onboardingCompletedProvider = Provider<bool>((ref) {
  final prefs = ref.watch(_sharedPrefsProvider);
  return prefs.getBool('onboarding_completed') ?? false;
});

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final prefs = await SharedPreferences.getInstance();
  runApp(
    ProviderScope(
      overrides: [
        _sharedPrefsProvider.overrideWithValue(prefs),
      ],
      child: PhysicsCopilotApp(prefs: prefs),
    ),
  );
}

// ---------------------------------------------------------------------------
// App widget
// ---------------------------------------------------------------------------

class PhysicsCopilotApp extends ConsumerWidget {
  const PhysicsCopilotApp({super.key, required this.prefs});

  final SharedPreferences prefs;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final onboardingDone = ref.watch(onboardingCompletedProvider);
    final selectedPrinter = ref.watch(printerProvider);

    final router = GoRouter(
      initialLocation: '/splash',
      redirect: (BuildContext ctx, GoRouterState state) {
        final location = state.matchedLocation;

        // Allow the splash screen to render freely.
        if (location == '/splash') return null;

        // 1. Onboarding not completed → redirect there (unless already there).
        if (!onboardingDone) {
          return location == '/onboarding' ? null : '/onboarding';
        }

        // 2. No printer selected → redirect to printer selection.
        if (selectedPrinter == null) {
          return location == '/printer-select' ? null : '/printer-select';
        }

        // 3. Everything is set up — allow navigation freely.
        return null;
      },
      routes: [
        GoRoute(
          path: '/splash',
          builder: (_, __) => const _SplashScreen(),
        ),
        GoRoute(
          path: '/onboarding',
          builder: (ctx, __) => OnboardingScreen(
            onComplete: () => ctx.go('/printer-select'),
          ),
        ),
        GoRoute(
          path: '/printer-select',
          builder: (ctx, __) => PrinterSelectionScreen(
            onComplete: () => ctx.go('/home'),
          ),
        ),
        GoRoute(
          path: '/home',
          builder: (ctx, __) => HomeScreen(
            onChangePrinter: () => ctx.push('/printer-select'),
            onStartCamera: () => ctx.push('/camera'),
          ),
        ),
        GoRoute(
          path: '/camera',
          builder: (_, __) => const CameraScreen(),
        ),
        GoRoute(
          path: '/history',
          builder: (_, __) => const HistoryScreen(),
        ),
      ],
    );

    return MaterialApp.router(
      title: 'PhysicsCopilot',
      debugShowCheckedModeBanner: false,
      theme: _buildDarkTheme(),
      routerConfig: router,
    );
  }

  ThemeData _buildDarkTheme() => ThemeData(
        colorScheme: const ColorScheme.dark(
          primary: Color(0xFF1B4F72),
          secondary: Color(0xFF2E86C1),
          surface: Color(0xFF1A1A2E),
        ),
        scaffoldBackgroundColor: const Color(0xFF0D0D1A),
        textTheme: GoogleFonts.poppinsTextTheme(ThemeData.dark().textTheme),
        useMaterial3: true,
      );
}

// ---------------------------------------------------------------------------
// Splash screen
// ---------------------------------------------------------------------------

class _SplashScreen extends StatelessWidget {
  const _SplashScreen();

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.science_rounded,
              size: 72,
              color: Color(0xFF2E86C1),
            ),
            SizedBox(height: 24),
            Text(
              'PhysicsCopilot',
              style: TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
                color: Colors.white,
                letterSpacing: 1.2,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
