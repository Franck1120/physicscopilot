import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:permission_handler/permission_handler.dart';

import '../main.dart' show onboardingCompletedProvider;
import '../providers/prefs_provider.dart' show sharedPrefsProvider;

// ---------------------------------------------------------------------------
// Data model for each onboarding slide
// ---------------------------------------------------------------------------

class _PageData {
  final IconData icon;
  final Color color;
  final String title;
  final String subtitle;
  final String detail;

  const _PageData({
    required this.icon,
    required this.color,
    required this.title,
    required this.subtitle,
    required this.detail,
  });
}

const List<_PageData> _pages = [
  _PageData(
    icon: Icons.camera_alt_rounded,
    color: Color(0xFF10B981), // emerald
    title: 'Punta la camera',
    subtitle: 'PhysicsCopilot vede il dispositivo in tempo reale',
    detail:
        'Inquadra il problema: errori, cavi, componenti rotti — l\'AI vede tutto.',
  ),
  _PageData(
    icon: Icons.auto_fix_high_rounded,
    color: Color(0xFF6366F1), // indigo
    title: 'L\'AI analizza',
    subtitle: 'Gemini identifica il problema in secondi',
    detail:
        'Riconoscimento visivo + knowledge base di riparazione per una diagnosi precisa.',
  ),
  _PageData(
    icon: Icons.check_circle_rounded,
    color: Color(0xFF10B981), // emerald
    title: 'Segui le istruzioni',
    subtitle: 'Guida passo-passo vocale e visiva',
    detail: 'Ogni step è verificato. Se hai dubbi, chiedi all\'AI direttamente.',
  ),
];

// ---------------------------------------------------------------------------
// Main screen — logic unchanged, UI enhanced
// ---------------------------------------------------------------------------

/// Onboarding flow shown once on first launch.
/// Calls [onComplete] after saving the completion flag so the router
/// can navigate to the next screen.
class OnboardingScreen extends ConsumerStatefulWidget {
  final VoidCallback onComplete;

  const OnboardingScreen({super.key, required this.onComplete});

  @override
  ConsumerState<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends ConsumerState<OnboardingScreen> {
  static const String _prefKey = 'onboarding_completed';
  static const int _pageCount = 3;

  final PageController _pageController = PageController();
  int _currentPage = 0;
  bool _isCompleting = false;

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  void _goToNextPage() {
    _pageController.nextPage(
      duration: const Duration(milliseconds: 350),
      curve: Curves.easeInOut,
    );
  }

  Future<void> _handleComplete() async {
    if (_isCompleting) return;
    setState(() => _isCompleting = true);

    await _requestPermissions();
    await _saveOnboardingCompleted();

    widget.onComplete();
  }

  Future<void> _requestPermissions() async {
    final statuses = await [
      Permission.camera,
      Permission.microphone,
    ].request();

    final cameraDenied = statuses[Permission.camera]?.isDenied ?? true;
    final micDenied = statuses[Permission.microphone]?.isDenied ?? true;

    if ((cameraDenied || micDenied) && mounted) {
      final denied = <String>[
        if (cameraDenied) 'camera',
        if (micDenied) 'microfono',
      ];
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Permessi ${denied.join(" e ")} negati. '
            'Alcune funzioni potrebbero non funzionare.',
          ),
          duration: const Duration(seconds: 4),
          behavior: SnackBarBehavior.floating,
        ),
      );
    }
  }

  /// Persists the completion flag to SharedPreferences and immediately
  /// updates the Riverpod [StateProvider] so the GoRouter redirect guard
  /// sees the new value without waiting for a provider recomputation.
  Future<void> _saveOnboardingCompleted() async {
    final prefs = ref.read(sharedPrefsProvider);
    await prefs.setBool(_prefKey, true);
    ref.read(onboardingCompletedProvider.notifier).state = true;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          // Base gradient background
          const DecoratedBox(
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [Color(0xFF1B4F72), Color(0xFF000000)],
              ),
            ),
            child: SizedBox.expand(),
          ),
          // Subtle geometric overlay — concentric arcs in top-right corner
          const Positioned.fill(
            child: CustomPaint(painter: _BackgroundPainter()),
          ),
          // Main content
          SafeArea(
            child: Column(
              children: [
                Expanded(
                  child: PageView(
                    controller: _pageController,
                    physics: const BouncingScrollPhysics(),
                    onPageChanged: (index) =>
                        setState(() => _currentPage = index),
                    children: _pages
                        .map(
                          (data) => _OnboardingPage(
                            key: ValueKey(data.title),
                            data: data,
                          ),
                        )
                        .toList(),
                  ),
                ),
                _PageIndicator(
                  pageCount: _pageCount,
                  currentPage: _currentPage,
                ),
                const SizedBox(height: 32),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 40),
                  child: AnimatedSwitcher(
                    duration: const Duration(milliseconds: 250),
                    transitionBuilder: (child, animation) => FadeTransition(
                      opacity: animation,
                      child: SlideTransition(
                        position: Tween<Offset>(
                          begin: const Offset(0, 0.15),
                          end: Offset.zero,
                        ).animate(animation),
                        child: child,
                      ),
                    ),
                    child: Semantics(
                      label: _currentPage < _pageCount - 1
                          ? 'Avanti alla slide successiva'
                          : 'Inizia a usare l\'app',
                      button: true,
                      child: _currentPage < _pageCount - 1
                          ? _PrimaryButton(
                              key: const ValueKey('avanti'),
                              label: 'Avanti',
                              onPressed: _goToNextPage,
                            )
                          : _PrimaryButton(
                              key: const ValueKey('inizia'),
                              label: 'Inizia',
                              onPressed: _isCompleting ? null : _handleComplete,
                            ),
                    ),
                  ),
                ),
                const SizedBox(height: 48),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Animated slide page
// ---------------------------------------------------------------------------

class _OnboardingPage extends StatefulWidget {
  final _PageData data;

  const _OnboardingPage({super.key, required this.data});

  @override
  State<_OnboardingPage> createState() => _OnboardingPageState();
}

class _OnboardingPageState extends State<_OnboardingPage>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;
  late final Animation<double> _fade;
  late final Animation<Offset> _slide;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 600),
    );
    _fade = CurvedAnimation(parent: _ctrl, curve: Curves.easeOut);
    _slide = Tween<Offset>(
      begin: const Offset(0, 0.12),
      end: Offset.zero,
    ).animate(CurvedAnimation(parent: _ctrl, curve: Curves.easeOutCubic));
    _ctrl.forward();
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FadeTransition(
      opacity: _fade,
      child: SlideTransition(
        position: _slide,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // Illustrated icon container — decorative, excluded from screen reader
              ExcludeSemantics(
                child: Container(
                  width: 140,
                  height: 140,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: widget.data.color.withAlpha(25),
                    border: Border.all(
                      color: widget.data.color.withAlpha(60),
                      width: 1.5,
                    ),
                  ),
                  child: Icon(
                    widget.data.icon,
                    size: 64,
                    color: widget.data.color,
                  ),
                ),
              ),
              const SizedBox(height: 40),
              Text(
                widget.data.title,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 26,
                  fontWeight: FontWeight.bold,
                  letterSpacing: 0.3,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 16),
              Text(
                widget.data.subtitle,
                style: TextStyle(
                  color: Colors.white.withValues(alpha: 0.7),
                  fontSize: 15,
                  height: 1.6,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 12),
              Text(
                widget.data.detail,
                style: TextStyle(
                  color: Colors.white.withValues(alpha: 0.45),
                  fontSize: 13,
                  height: 1.5,
                ),
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
// Background geometric painter
// ---------------------------------------------------------------------------

class _BackgroundPainter extends CustomPainter {
  const _BackgroundPainter();

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = const Color(0xFF10B981).withAlpha(15)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;
    for (int i = 1; i <= 4; i++) {
      canvas.drawCircle(
        Offset(size.width, 0),
        i * 80.0,
        paint,
      );
    }
  }

  @override
  bool shouldRepaint(_BackgroundPainter _) => false;
}

// ---------------------------------------------------------------------------
// Page indicator — unchanged
// ---------------------------------------------------------------------------

class _PageIndicator extends StatelessWidget {
  final int pageCount;
  final int currentPage;

  const _PageIndicator({
    required this.pageCount,
    required this.currentPage,
  });

  @override
  Widget build(BuildContext context) {
    return Semantics(
      label: 'Slide ${currentPage + 1} di $pageCount',
      child: ExcludeSemantics(
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: List.generate(pageCount, (index) {
            final isActive = index == currentPage;
            return AnimatedContainer(
              duration: const Duration(milliseconds: 250),
              curve: Curves.easeInOut,
              margin: const EdgeInsets.symmetric(horizontal: 4),
              width: isActive ? 24 : 8,
              height: 8,
              decoration: BoxDecoration(
                color: isActive ? Colors.white : const Color(0x66FFFFFF),
                borderRadius: BorderRadius.circular(4),
              ),
            );
          }),
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Primary button — unchanged
// ---------------------------------------------------------------------------

class _PrimaryButton extends StatelessWidget {
  final String label;
  final VoidCallback? onPressed;

  const _PrimaryButton({
    super.key,
    required this.label,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      height: 52,
      child: ElevatedButton(
        onPressed: onPressed,
        style: ElevatedButton.styleFrom(
          backgroundColor: Colors.white,
          foregroundColor: const Color(0xFF1B4F72),
          disabledBackgroundColor: const Color(0x66FFFFFF),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(26),
          ),
          elevation: 0,
          textStyle: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.5,
          ),
        ),
        child: Text(label),
      ),
    );
  }
}
