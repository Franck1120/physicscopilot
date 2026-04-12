import 'package:flutter/material.dart';

/// Creates a [PageRouteBuilder] that slides the new page in from the right.
///
/// Uses a 300 ms [CurvedAnimation] with [Curves.easeInOut] for a polished feel.
PageRouteBuilder<T> slideFromRight<T>(Widget page) {
  return PageRouteBuilder<T>(
    pageBuilder: (context, animation, secondaryAnimation) => page,
    transitionDuration: const Duration(milliseconds: 300),
    reverseTransitionDuration: const Duration(milliseconds: 250),
    transitionsBuilder: (context, animation, secondaryAnimation, child) {
      final tween = Tween(
        begin: const Offset(1.0, 0.0),
        end: Offset.zero,
      ).chain(CurveTween(curve: Curves.easeInOut));

      return SlideTransition(
        position: animation.drive(tween),
        child: child,
      );
    },
  );
}
