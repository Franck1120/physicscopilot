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
// Design tokens
// ---------------------------------------------------------------------------

/// Brand accent — emerald green, consistent with landing page palette.
const Color kAccent = Color(0xFF10B981);
const Color kAccentDark = Color(0xFF059669);
const Color kBgPrimary = Color(0xFF0A0A0A);
const Color kBgSurface = Color(0xFF141414);
const Color kBgCard = Color(0xFF1A1A1A);
const Color kBgCardBorder = Color(0xFF2A2A2A);
const Color kTextMuted = Color(0xFF6B7280);

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
          pageBuilder: (_, state) => _fadePage(state, const _SplashScreen()),
        ),
        GoRoute(
          path: '/onboarding',
          pageBuilder: (ctx, state) => _fadePage(
            state,
            OnboardingScreen(onComplete: () => ctx.go('/printer-select')),
          ),
        ),
        GoRoute(
          path: '/printer-select',
          pageBuilder: (ctx, state) => _fadePage(
            state,
            PrinterSelectionScreen(onComplete: () => ctx.go('/home')),
          ),
        ),
        GoRoute(
          path: '/home',
          pageBuilder: (ctx, state) => _fadePage(
            state,
            HomeScreen(
              onChangePrinter: () => ctx.push('/printer-select'),
              onStartCamera: () => ctx.push('/camera'),
            ),
          ),
        ),
        GoRoute(
          path: '/camera',
          pageBuilder: (_, state) => _fadePage(state, const CameraScreen()),
        ),
        GoRoute(
          path: '/history',
          pageBuilder: (_, state) => _fadePage(state, const HistoryScreen()),
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

  /// Wraps a page in a cross-fade transition.
  static CustomTransitionPage<void> _fadePage(
    GoRouterState state,
    Widget child,
  ) {
    return CustomTransitionPage<void>(
      key: state.pageKey,
      child: child,
      transitionDuration: const Duration(milliseconds: 280),
      transitionsBuilder: (_, animation, __, child) => FadeTransition(
        opacity: CurvedAnimation(parent: animation, curve: Curves.easeIn),
        child: child,
      ),
    );
  }

  ThemeData _buildDarkTheme() => ThemeData(
        colorScheme: ColorScheme.dark(
          primary: kAccent,
          secondary: kAccentDark,
          surface: kBgSurface,
        ),
        scaffoldBackgroundColor: kBgPrimary,
        textTheme: GoogleFonts.poppinsTextTheme(ThemeData.dark().textTheme),
        useMaterial3: true,
      );
}

// ---------------------------------------------------------------------------
// Splash screen — animated logo reveal, auto-navigates after 2.2 s
// ---------------------------------------------------------------------------

class _SplashScreen extends StatefulWidget {
  const _SplashScreen();

  @override
  State<_SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<_SplashScreen>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;
  late final Animation<double> _fade;
  late final Animation<double> _scale;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    );
    _fade = CurvedAnimation(parent: _ctrl, curve: Curves.easeIn);
    _scale = Tween<double>(begin: 0.80, end: 1.0).animate(
      CurvedAnimation(parent: _ctrl, curve: Curves.easeOutCubic),
    );
    _ctrl.forward();
    Future.delayed(const Duration(milliseconds: 2200), () {
      if (mounted) context.go('/home');
    });
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: kBgPrimary,
      body: Center(
        child: FadeTransition(
          opacity: _fade,
          child: ScaleTransition(
            scale: _scale,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Logo with emerald glow
                Container(
                  width: 100,
                  height: 100,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: kAccent.withAlpha(20),
                    boxShadow: [
                      BoxShadow(
                        color: kAccent.withAlpha(70),
                        blurRadius: 48,
                        spreadRadius: 12,
                      ),
                    ],
                  ),
                  child: const Icon(
                    Icons.science_rounded,
                    size: 52,
                    color: kAccent,
                  ),
                ),
                const SizedBox(height: 32),
                const Text(
                  'PhysicsCopilot',
                  style: TextStyle(
                    fontSize: 30,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                    letterSpacing: 1.5,
                  ),
                ),
                const SizedBox(height: 8),
                const Text(
                  'Your AI repair assistant',
                  style: TextStyle(
                    fontSize: 14,
                    color: kAccent,
                    letterSpacing: 0.8,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
