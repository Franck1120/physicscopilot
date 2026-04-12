import 'package:flutter/material.dart';

import '../main.dart' show kBgPrimary, kTextMuted, kAccent;

/// Returns a full-screen Scaffold that displays [error] in a graceful card.
///
/// Designed to be returned from a screen's `build()` method when a
/// synchronous exception is caught:
///
/// ```dart
/// @override
/// Widget build(BuildContext context) {
///   try {
///     return _buildContent(context);
///   } catch (e) {
///     return screenError(e, context);
///   }
/// }
/// ```
///
/// A "Torna indietro" button is shown when the navigator can pop.
/// The error message is truncated to 180 characters.
Widget screenError(Object error, BuildContext context) {
  final message = error.toString();
  final preview =
      message.length > 180 ? '${message.substring(0, 180)}…' : message;

  return Scaffold(
    backgroundColor: kBgPrimary,
    body: SafeArea(
      child: Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(
                Icons.warning_amber_rounded,
                color: Colors.orangeAccent,
                size: 52,
              ),
              const SizedBox(height: 20),
              const Text(
                'Qualcosa è andato storto',
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              Text(
                preview,
                style: const TextStyle(color: kTextMuted, fontSize: 12),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 28),
              if (Navigator.of(context).canPop())
                OutlinedButton.icon(
                  onPressed: () => Navigator.of(context).pop(),
                  icon: const Icon(Icons.arrow_back, size: 16),
                  label: const Text('Torna indietro'),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: Colors.white70,
                    side: const BorderSide(color: Colors.white24),
                  ),
                ),
              const SizedBox(height: 12),
              ElevatedButton.icon(
                onPressed: () {
                  // Navigate to home as a safe fallback.
                  Navigator.of(context).popUntil((route) => route.isFirst);
                },
                icon: const Icon(Icons.home_outlined, size: 16),
                label: const Text('Vai alla home'),
                style: ElevatedButton.styleFrom(
                  backgroundColor: kAccent,
                  foregroundColor: Colors.white,
                ),
              ),
            ],
          ),
        ),
      ),
    ),
  );
}
