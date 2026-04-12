import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:app_links/app_links.dart';
import 'providers/prefs_provider.dart';
import 'screens/onboarding_screen.dart';
import 'services/notification_service.dart';
import 'utils/strings.dart';
import 'screens/equipment_selection_screen.dart';
import 'screens/home_screen.dart';
import 'screens/camera_screen.dart';
import 'screens/session_screen.dart';
import 'screens/history_screen.dart';
import 'screens/settings_screen.dart';
import 'providers/equipment_provider.dart';
import 'providers/settings_provider.dart';
import 'package:supabase_flutter/supabase_flutter.dart' show AuthState;
import 'screens/login_screen.dart';
import 'services/auth_service.dart';

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
// WCAG AA contrast (≥4.5:1) on kBgPrimary (#0A0A0A): #9CA3AF ≈ 7.8:1
const Color kTextMuted = Color(0xFF9CA3AF);

// ---------------------------------------------------------------------------
// Providers
// ---------------------------------------------------------------------------

/// True if the user has already completed onboarding.
///
/// [StateProvider] so that [OnboardingScreen] can update it directly via
/// `ref.read(onboardingCompletedProvider.notifier).state = true` without
/// needing to invalidate the cache manually — Riverpod propagates the change
/// to all listeners (including the GoRouter redirect guard) immediately.
final onboardingCompletedProvider = StateProvider<bool>((ref) {
  final prefs = ref.watch(sharedPrefsProvider);
  return prefs.getBool('onboarding_completed') ?? false;
});

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------

void main() async {
  final binding = WidgetsFlutterBinding.ensureInitialized();

  // Forward Flutter framework errors to the standard error log without crashing.
  FlutterError.onError = FlutterError.presentError;

  // Catch unhandled async / platform errors and absorb them gracefully.
  binding.platformDispatcher.onError = (error, stack) {
    debugPrint('[PhysicsCopilot] Unhandled error: $error\n$stack');
    return true;
  };

  // Replace Flutter's red-screen with a friendly dark error widget.
  ErrorWidget.builder = (FlutterErrorDetails details) =>
      _AppErrorWidget(message: details.exceptionAsString());

  await AuthService.initialize();
  await NotificationService.initialize();

  final prefs = await SharedPreferences.getInstance();
  runApp(
    ProviderScope(
      overrides: [
        sharedPrefsProvider.overrideWithValue(prefs),
      ],
      child: PhysicsCopilotApp(prefs: prefs),
    ),
  );
}

// ---------------------------------------------------------------------------
// App widget
// ---------------------------------------------------------------------------

/// [PhysicsCopilotApp] is a [ConsumerStatefulWidget] so that [GoRouter] is
/// created once in [initState] and stored as a field.
///
/// Previously [GoRouter] was instantiated inside [build], which recreated it
/// on every rebuild — resetting navigation state on provider changes.
/// [ref.listenManual] triggers [_router.refresh()] when the redirect-relevant
/// providers change, so the guard logic re-runs without recreating the router.
class PhysicsCopilotApp extends ConsumerStatefulWidget {
  const PhysicsCopilotApp({super.key, required this.prefs});

  final SharedPreferences prefs;

  @override
  ConsumerState<PhysicsCopilotApp> createState() => _PhysicsCopilotAppState();
}

class _PhysicsCopilotAppState extends ConsumerState<PhysicsCopilotApp> {
  late final GoRouter _router;
  late final ProviderSubscription<bool> _onboardingSub;
  late final ProviderSubscription<EquipmentProfile?> _equipmentSub;
  StreamSubscription<Uri>? _deepLinkSub;
  StreamSubscription<AuthState>? _authSub;

  @override
  void initState() {
    super.initState();

    _router = GoRouter(
      initialLocation: '/splash',
      redirect: (BuildContext ctx, GoRouterState state) {
        // Read current values; safe because refresh() is called on changes.
        final onboardingDone = ref.read(onboardingCompletedProvider);
        final selectedEquipment = ref.read(equipmentProvider);
        final location = state.matchedLocation;

        if (location == '/splash') return null;

        // Auth guard: redirect to /login when Supabase is configured and
        // the user is not signed in. Allow /login itself to always render.
        if (!AuthService.isAuthenticated && location != '/login') {
          // Only block when Supabase is actually configured (i.e. not dev mode).
          // Dev mode: SUPABASE_URL not set → isAuthenticated is always false
          // but we skip the gate so developers can run without credentials.
          const supabaseUrl = String.fromEnvironment('SUPABASE_URL');
          if (supabaseUrl.isNotEmpty) return '/login';
        }
        if (location == '/login') return null;

        if (!onboardingDone) {
          return location == '/onboarding' ? null : '/onboarding';
        }

        if (selectedEquipment == null) {
          return location == '/equipment-select' ? null : '/equipment-select';
        }

        return null;
      },
      routes: [
        GoRoute(
          path: '/login',
          pageBuilder: (_, state) => _fadePage(state, const LoginScreen()),
        ),
        GoRoute(
          path: '/splash',
          pageBuilder: (_, state) => _fadePage(state, const _SplashScreen()),
        ),
        GoRoute(
          path: '/onboarding',
          pageBuilder: (ctx, state) => _fadePage(
            state,
            OnboardingScreen(onComplete: () => ctx.go('/equipment-select')),
          ),
        ),
        GoRoute(
          path: '/equipment-select',
          pageBuilder: (ctx, state) => _fadePage(
            state,
            EquipmentSelectionScreen(onComplete: () => ctx.go('/home')),
          ),
        ),
        GoRoute(
          path: '/home',
          pageBuilder: (ctx, state) => _fadePage(
            state,
            HomeScreen(
              onChangeEquipment: () => ctx.push('/equipment-select'),
              onStartCamera: () => ctx.push('/session'),
            ),
          ),
        ),
        // Session and camera slide up from the bottom (modal feel).
        GoRoute(
          path: '/session',
          pageBuilder: (_, state) =>
              _slidePage(state, const SessionScreen(), fromBottom: true),
        ),
        GoRoute(
          path: '/camera',
          pageBuilder: (_, state) =>
              _slidePage(state, const CameraScreen(), fromBottom: true),
        ),
        // Secondary screens slide in from the right.
        GoRoute(
          path: '/history',
          pageBuilder: (_, state) =>
              _slidePage(state, const HistoryScreen()),
        ),
        GoRoute(
          path: '/settings',
          pageBuilder: (_, state) =>
              _slidePage(state, const SettingsScreen()),
        ),
      ],
    );

    // Re-run redirect guards when auth-related state changes.
    _onboardingSub = ref.listenManual(
      onboardingCompletedProvider,
      (prev, next) {
        if (prev != next) _router.refresh();
      },
    );
    _equipmentSub = ref.listenManual(
      equipmentProvider,
      (prev, next) {
        if (prev?.id != next?.id) _router.refresh();
      },
    );

    // Refresh router when auth state changes (sign-in / sign-out).
    _authSub = AuthService.authStateChanges.listen((_) => _router.refresh());

    // Handle deep links: physicscopilot://session/new → navigate to /session
    _deepLinkSub = AppLinks().uriLinkStream.listen(_handleDeepLink);
  }

  void _handleDeepLink(Uri uri) {
    if (uri.scheme != 'physicscopilot') return;
    final path = uri.pathSegments;
    if (path.isEmpty) return;
    switch (path.first) {
      case 'session':
        _router.go('/session');
      case 'history':
        _router.go('/history');
      default:
        break;
    }
  }

  @override
  void dispose() {
    _authSub?.cancel();
    _deepLinkSub?.cancel();
    _onboardingSub.close();
    _equipmentSub.close();
    _router.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final themeMode = ref.watch(
      settingsProvider.select((s) => s.themeMode),
    );
    return MaterialApp.router(
      title: AppStrings.appName,
      debugShowCheckedModeBanner: false,
      theme: _buildLightTheme(),
      darkTheme: _buildDarkTheme(),
      themeMode: themeMode,
      routerConfig: _router,
    );
  }

  /// Cross-fade transition — used for root/primary screens.
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

  /// Slide transition — right-to-left for secondary screens, bottom-to-top
  /// for full-screen overlays like session and camera.
  static CustomTransitionPage<void> _slidePage(
    GoRouterState state,
    Widget child, {
    bool fromBottom = false,
  }) {
    return CustomTransitionPage<void>(
      key: state.pageKey,
      child: child,
      transitionDuration: const Duration(milliseconds: 340),
      reverseTransitionDuration: const Duration(milliseconds: 260),
      transitionsBuilder: (_, animation, __, child) {
        final begin =
            fromBottom ? const Offset(0.0, 1.0) : const Offset(1.0, 0.0);
        final tween = Tween<Offset>(begin: begin, end: Offset.zero)
            .chain(CurveTween(curve: Curves.easeOutCubic));
        final reverseTween = Tween<Offset>(begin: begin, end: Offset.zero)
            .chain(CurveTween(curve: Curves.easeInCubic));
        final slide = animation.status == AnimationStatus.reverse
            ? animation.drive(reverseTween)
            : animation.drive(tween);
        return SlideTransition(
          position: slide,
          child: child,
        );
      },
    );
  }

  ThemeData _buildLightTheme() => ThemeData(
        colorScheme: ColorScheme.light(
          primary: kAccent,
          secondary: kAccentDark,
          surface: const Color(0xFFF9FAFB),
          onSurface: const Color(0xFF111827),
          onPrimary: Colors.white,
          error: Colors.redAccent,
        ),
        scaffoldBackgroundColor: const Color(0xFFF3F4F6),
        appBarTheme: const AppBarTheme(
          backgroundColor: Colors.white,
          foregroundColor: Color(0xFF111827),
          elevation: 0,
          centerTitle: false,
          iconTheme: IconThemeData(color: Color(0xFF111827)),
          titleTextStyle: TextStyle(
            color: Color(0xFF111827),
            fontSize: 18,
            fontWeight: FontWeight.bold,
            letterSpacing: 0.4,
          ),
        ),
        cardTheme: CardThemeData(
          color: Colors.white,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: const BorderSide(color: Color(0xFFE5E7EB), width: 1),
          ),
        ),
        bottomNavigationBarTheme: const BottomNavigationBarThemeData(
          backgroundColor: Colors.white,
          selectedItemColor: kAccent,
          unselectedItemColor: Color(0xFF9CA3AF),
          type: BottomNavigationBarType.fixed,
          elevation: 0,
        ),
        snackBarTheme: SnackBarThemeData(
          backgroundColor: const Color(0xFF1F2937),
          contentTextStyle:
              const TextStyle(color: Colors.white, fontSize: 14),
          shape:
              RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          behavior: SnackBarBehavior.floating,
          actionTextColor: kAccent,
        ),
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: kAccent,
            foregroundColor: Colors.white,
            elevation: 0,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        ),
        outlinedButtonTheme: OutlinedButtonThemeData(
          style: OutlinedButton.styleFrom(
            foregroundColor: kAccent,
            side: const BorderSide(color: Color(0xFFD1D5DB)),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        ),
        switchTheme: SwitchThemeData(
          thumbColor: WidgetStateProperty.resolveWith(
            (states) =>
                states.contains(WidgetState.selected) ? kAccent : null,
          ),
          trackColor: WidgetStateProperty.resolveWith(
            (states) => states.contains(WidgetState.selected)
                ? kAccent.withAlpha(80)
                : null,
          ),
        ),
        dialogTheme: DialogThemeData(
          backgroundColor: Colors.white,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
            side: const BorderSide(color: Color(0xFFE5E7EB), width: 1),
          ),
          titleTextStyle: const TextStyle(
            color: Color(0xFF111827),
            fontSize: 18,
            fontWeight: FontWeight.bold,
          ),
          contentTextStyle: const TextStyle(
              color: Color(0xFF374151), fontSize: 14, height: 1.5),
        ),
        textTheme: GoogleFonts.poppinsTextTheme(ThemeData.light().textTheme),
        useMaterial3: true,
      );

  ThemeData _buildDarkTheme() => ThemeData(
        colorScheme: ColorScheme.dark(
          primary: kAccent,
          secondary: kAccentDark,
          surface: kBgSurface,
          onSurface: Colors.white,
          onPrimary: Colors.white,
          error: Colors.redAccent,
        ),
        scaffoldBackgroundColor: kBgPrimary,

        // AppBar — consistent dark header across all screens.
        appBarTheme: const AppBarTheme(
          backgroundColor: Color(0xFF111111),
          foregroundColor: Colors.white,
          elevation: 0,
          centerTitle: false,
          iconTheme: IconThemeData(color: Colors.white),
          titleTextStyle: TextStyle(
            color: Colors.white,
            fontSize: 18,
            fontWeight: FontWeight.bold,
            letterSpacing: 0.4,
          ),
        ),

        // Card — matches kBgCard with rounded corners and a subtle border.
        cardTheme: CardThemeData(
          color: kBgCard,
          elevation: 0,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
            side: const BorderSide(color: kBgCardBorder, width: 1),
          ),
        ),

        // BottomNavigationBar
        bottomNavigationBarTheme: const BottomNavigationBarThemeData(
          backgroundColor: Color(0xFF111111),
          selectedItemColor: kAccent,
          unselectedItemColor: kTextMuted,
          type: BottomNavigationBarType.fixed,
          elevation: 0,
        ),

        // SnackBar
        snackBarTheme: SnackBarThemeData(
          backgroundColor: kBgCard,
          contentTextStyle: const TextStyle(color: Colors.white, fontSize: 14),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          behavior: SnackBarBehavior.floating,
          actionTextColor: kAccent,
        ),

        // ElevatedButton — emerald fill.
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: kAccent,
            foregroundColor: Colors.white,
            elevation: 0,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        ),

        // OutlinedButton
        outlinedButtonTheme: OutlinedButtonThemeData(
          style: OutlinedButton.styleFrom(
            foregroundColor: kAccent,
            side: const BorderSide(color: kBgCardBorder),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(8),
            ),
          ),
        ),

        // Switch — emerald thumb when active.
        switchTheme: SwitchThemeData(
          thumbColor: WidgetStateProperty.resolveWith(
            (states) =>
                states.contains(WidgetState.selected) ? kAccent : null,
          ),
          trackColor: WidgetStateProperty.resolveWith(
            (states) =>
                states.contains(WidgetState.selected)
                    ? kAccent.withAlpha(80)
                    : null,
          ),
        ),

        // Dialog
        dialogTheme: DialogThemeData(
          backgroundColor: kBgCard,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
            side: const BorderSide(color: kBgCardBorder, width: 1),
          ),
          titleTextStyle: const TextStyle(
            color: Colors.white,
            fontSize: 18,
            fontWeight: FontWeight.bold,
          ),
          contentTextStyle:
              const TextStyle(color: Colors.white70, fontSize: 14, height: 1.5),
        ),

        textTheme: GoogleFonts.poppinsTextTheme(ThemeData.dark().textTheme),
        useMaterial3: true,
      );
}

// ---------------------------------------------------------------------------
// Global error widget — replaces Flutter's red screen on widget build errors
// ---------------------------------------------------------------------------

class _AppErrorWidget extends StatelessWidget {
  const _AppErrorWidget({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: kBgPrimary,
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(
                Icons.warning_amber_rounded,
                color: Colors.orangeAccent,
                size: 52,
              ),
              const SizedBox(height: 20),
              const Text(
                AppStrings.errorGeneric,
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              const Text(
                AppStrings.restartApp,
                style: TextStyle(color: kTextMuted, fontSize: 14),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Splash screen — staggered animated reveal, auto-navigates after 2.4 s
// ---------------------------------------------------------------------------

class _SplashScreen extends StatefulWidget {
  const _SplashScreen();

  @override
  State<_SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<_SplashScreen>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;

  // Logo: scale + fade — leads the sequence
  late final Animation<double> _logoScale;
  late final Animation<double> _logoFade;

  // Title: slides up + fades after logo
  late final Animation<Offset> _titleSlide;
  late final Animation<double> _titleFade;

  // Subtitle + loading dot fade in last
  late final Animation<double> _subtitleFade;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1600),
    );

    _logoScale = Tween<double>(begin: 0.60, end: 1.0).animate(
      CurvedAnimation(
        parent: _ctrl,
        curve: const Interval(0.0, 0.55, curve: Curves.easeOutBack),
      ),
    );
    _logoFade = CurvedAnimation(
      parent: _ctrl,
      curve: const Interval(0.0, 0.45, curve: Curves.easeIn),
    );
    _titleSlide = Tween<Offset>(
      begin: const Offset(0, 0.4),
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _ctrl,
      curve: const Interval(0.35, 0.70, curve: Curves.easeOut),
    ));
    _titleFade = CurvedAnimation(
      parent: _ctrl,
      curve: const Interval(0.35, 0.65, curve: Curves.easeIn),
    );
    _subtitleFade = CurvedAnimation(
      parent: _ctrl,
      curve: const Interval(0.60, 0.90, curve: Curves.easeIn),
    );

    _ctrl.forward();
    Future.delayed(const Duration(milliseconds: 2400), () {
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
      body: Stack(
        alignment: Alignment.center,
        children: [
          // Centre content
          Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // Logo
              ScaleTransition(
                scale: _logoScale,
                child: FadeTransition(
                  opacity: _logoFade,
                  child: Container(
                    width: 108,
                    height: 108,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      color: kAccent.withAlpha(18),
                      border: Border.all(
                        color: kAccent.withAlpha(60),
                        width: 1.5,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: kAccent.withAlpha(80),
                          blurRadius: 56,
                          spreadRadius: 16,
                        ),
                      ],
                    ),
                    child: const Icon(
                      Icons.science_rounded,
                      size: 54,
                      color: kAccent,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 36),

              // Title
              SlideTransition(
                position: _titleSlide,
                child: FadeTransition(
                  opacity: _titleFade,
                  child: const Text(
                    AppStrings.appName,
                    style: TextStyle(
                      fontSize: 32,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: 1.2,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 10),

              // Subtitle
              FadeTransition(
                opacity: _subtitleFade,
                child: const Text(
                  AppStrings.appTagline,
                  style: TextStyle(
                    fontSize: 14,
                    color: kAccent,
                    letterSpacing: 0.5,
                  ),
                ),
              ),
            ],
          ),

          // Bottom loading indicator
          Positioned(
            bottom: 56,
            child: FadeTransition(
              opacity: _subtitleFade,
              child: const SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: kAccent,
                ),
              ),
            ),
          ),

          // Version label
          Positioned(
            bottom: 24,
            child: FadeTransition(
              opacity: _subtitleFade,
              child: const Text(
                AppStrings.appVersion,
                style: TextStyle(
                  fontSize: 11,
                  color: kTextMuted,
                  letterSpacing: 0.5,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
