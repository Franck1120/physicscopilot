// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:supabase_flutter/supabase_flutter.dart';

import '../services/auth_service.dart';

/// Email + password sign-in / sign-up screen.
///
/// On successful authentication GoRouter's redirect guard automatically
/// navigates the user to /home. Design follows the app's dark/emerald theme.
class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  bool _isSignUp = false;
  bool _loading = false;
  bool _obscurePassword = true;

  @override
  void dispose() {
    _emailCtrl.dispose();
    _passwordCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _loading = true);
    try {
      if (_isSignUp) {
        await AuthService.signUpWithEmail(
            _emailCtrl.text.trim(), _passwordCtrl.text);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
                content: Text('Account creato. Controlla la tua email.')),
          );
        }
      } else {
        await AuthService.signInWithEmail(
            _emailCtrl.text.trim(), _passwordCtrl.text);
        // GoRouter redirect guard re-runs and navigates to /home.
        if (mounted) context.go('/home');
      }
    } on AuthException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(e.message)));
      }
    } catch (_) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Errore di rete. Riprova.')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    const kAccent = Color(0xFF10B981);
    const kBgPrimary = Color(0xFF0A0A0A);
    const kBgCard = Color(0xFF1A1A1A);
    const kBgCardBorder = Color(0xFF2A2A2A);
    const kTextMuted = Color(0xFF9CA3AF);

    return Scaffold(
      backgroundColor: kBgPrimary,
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 420),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    const Icon(Icons.science_rounded, size: 56, color: kAccent),
                    const SizedBox(height: 16),
                    Text(
                      _isSignUp ? 'Crea account' : 'Accedi',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 28,
                        fontWeight: FontWeight.bold,
                        letterSpacing: 0.4,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      _isSignUp
                          ? 'Inizia ad usare PhysicsCopilot'
                          : 'Bentornato su PhysicsCopilot',
                      style: const TextStyle(color: kTextMuted, fontSize: 14),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 40),
                    // Email field
                    _buildInputContainer(
                      kBgCard: kBgCard,
                      kBgCardBorder: kBgCardBorder,
                      child: TextFormField(
                        controller: _emailCtrl,
                        keyboardType: TextInputType.emailAddress,
                        autofillHints: const [AutofillHints.email],
                        style: const TextStyle(color: Colors.white),
                        decoration: InputDecoration(
                          labelText: 'Email',
                          labelStyle: const TextStyle(color: kTextMuted),
                          prefixIcon:
                              const Icon(Icons.email_outlined, color: kTextMuted),
                          border: InputBorder.none,
                          contentPadding: const EdgeInsets.symmetric(
                              horizontal: 16, vertical: 14),
                        ),
                        validator: (v) {
                          if (v == null || v.trim().isEmpty) {
                            return 'Inserisci la tua email';
                          }
                          if (!v.contains('@')) return 'Email non valida';
                          return null;
                        },
                      ),
                    ),
                    const SizedBox(height: 12),
                    // Password field
                    _PasswordField(
                      ctrl: _passwordCtrl,
                      isSignUp: _isSignUp,
                      obscure: _obscurePassword,
                      onToggle: () =>
                          setState(() => _obscurePassword = !_obscurePassword),
                      onSubmit: _submit,
                    ),
                    const SizedBox(height: 24),
                    // Submit button
                    SizedBox(
                      height: 48,
                      child: ElevatedButton(
                        onPressed: _loading ? null : _submit,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: kAccent,
                          foregroundColor: Colors.white,
                          disabledBackgroundColor: kAccent.withAlpha(100),
                          elevation: 0,
                          shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(10)),
                        ),
                        child: _loading
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(
                                    strokeWidth: 2, color: Colors.white),
                              )
                            : Text(
                                _isSignUp ? 'Registrati' : 'Accedi',
                                style: const TextStyle(
                                    fontSize: 15,
                                    fontWeight: FontWeight.w600),
                              ),
                      ),
                    ),
                    const SizedBox(height: 20),
                    // Toggle sign-in / sign-up
                    TextButton(
                      onPressed: _loading
                          ? null
                          : () => setState(() => _isSignUp = !_isSignUp),
                      child: Text(
                        _isSignUp
                            ? 'Hai gia un account? Accedi'
                            : 'Non hai un account? Registrati',
                        style: const TextStyle(color: kAccent, fontSize: 14),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  static Widget _buildInputContainer({
    required Color kBgCard,
    required Color kBgCardBorder,
    required Widget child,
  }) {
    return Container(
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: kBgCardBorder),
      ),
      child: child,
    );
  }
}

class _PasswordField extends StatelessWidget {
  const _PasswordField({
    required this.ctrl,
    required this.isSignUp,
    required this.obscure,
    required this.onToggle,
    required this.onSubmit,
  });

  final TextEditingController ctrl;
  final bool isSignUp;
  final bool obscure;
  final VoidCallback onToggle;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    const kBgCard = Color(0xFF1A1A1A);
    const kBgCardBorder = Color(0xFF2A2A2A);
    const kTextMuted = Color(0xFF9CA3AF);
    return Container(
      decoration: BoxDecoration(
        color: kBgCard,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: kBgCardBorder),
      ),
      child: TextFormField(
        controller: ctrl,
        obscureText: obscure,
        autofillHints: isSignUp
            ? const [AutofillHints.newPassword]
            : const [AutofillHints.password],
        style: const TextStyle(color: Colors.white),
        decoration: InputDecoration(
          labelText: 'Password',
          labelStyle: const TextStyle(color: kTextMuted),
          prefixIcon: const Icon(Icons.lock_outline, color: kTextMuted),
          suffixIcon: IconButton(
            icon: Icon(
              obscure
                  ? Icons.visibility_off_outlined
                  : Icons.visibility_outlined,
              color: kTextMuted,
              size: 20,
            ),
            onPressed: onToggle,
          ),
          border: InputBorder.none,
          contentPadding:
              const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        ),
        validator: (v) {
          if (v == null || v.isEmpty) return 'Inserisci la tua password';
          if (isSignUp && v.length < 8) return 'Minimo 8 caratteri';
          return null;
        },
        onFieldSubmitted: (_) => onSubmit(),
      ),
    );
  }
}
