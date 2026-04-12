import 'package:flutter/material.dart';

import '../models/session_record.dart';
import '../utils/strings.dart';

// ── SessionBadge ─────────────────────────────────────────────────────────────

/// Small pill badge indicating whether a session was resolved or not.
///
/// Uses a green dot for [SessionStatus.resolved] and a red dot for
/// [SessionStatus.unresolved], matching the history screen colour scheme.
class SessionBadge extends StatelessWidget {
  const SessionBadge({super.key, required this.status});

  final SessionStatus status;

  @override
  Widget build(BuildContext context) {
    final isResolved = status == SessionStatus.resolved;

    final dotColor = isResolved ? const Color(0xFF10B981) : Colors.redAccent;
    final bgColor = isResolved
        ? const Color(0xFF10B981).withAlpha(26)
        : Colors.redAccent.withAlpha(26);
    final label = isResolved
        ? AppStrings.historyStatusResolved
        : AppStrings.historyStatusUnresolved;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 7,
            height: 7,
            decoration: BoxDecoration(color: dotColor, shape: BoxShape.circle),
          ),
          const SizedBox(width: 5),
          Text(
            label,
            style: TextStyle(
              color: dotColor,
              fontSize: 11,
              fontWeight: FontWeight.w600,
              letterSpacing: 0.2,
            ),
          ),
        ],
      ),
    );
  }
}
