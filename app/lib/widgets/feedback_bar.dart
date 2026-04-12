import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../main.dart' show kAccent, kTextMuted;
import '../providers/prefs_provider.dart';

/// Thumbs-up / thumbs-down buttons shown once the typewriter animation ends.
/// Selection is persisted to SharedPreferences as aggregate counters.
class FeedbackBar extends ConsumerStatefulWidget {
  const FeedbackBar({super.key, required this.responseText});
  final String responseText;

  @override
  ConsumerState<FeedbackBar> createState() => _FeedbackBarState();
}

class _FeedbackBarState extends ConsumerState<FeedbackBar> {
  // null = not voted, true = liked, false = disliked
  bool? _vote;

  static const _keyUp = 'feedback_thumbsUp';
  static const _keyDown = 'feedback_thumbsDown';

  Future<void> _setVote(bool liked) async {
    if (_vote != null) return; // already voted
    HapticFeedback.selectionClick();
    setState(() => _vote = liked);
    final prefs = ref.read(sharedPrefsProvider);
    if (liked) {
      await prefs.setInt(_keyUp, (prefs.getInt(_keyUp) ?? 0) + 1);
    } else {
      await prefs.setInt(_keyDown, (prefs.getInt(_keyDown) ?? 0) + 1);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Semantics(
          label: 'Risposta utile',
          button: true,
          child: IconButton(
            icon: Icon(
              _vote == true
                  ? Icons.thumb_up_rounded
                  : Icons.thumb_up_outlined,
              size: 15,
              color: _vote == true ? kAccent : kTextMuted,
            ),
            tooltip: 'Utile',
            onPressed: _vote == null ? () => _setVote(true) : null,
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
          ),
        ),
        Semantics(
          label: 'Risposta non utile',
          button: true,
          child: IconButton(
            icon: Icon(
              _vote == false
                  ? Icons.thumb_down_rounded
                  : Icons.thumb_down_outlined,
              size: 15,
              color: _vote == false ? Colors.redAccent : kTextMuted,
            ),
            tooltip: 'Non utile',
            onPressed: _vote == null ? () => _setVote(false) : null,
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
          ),
        ),
      ],
    );
  }
}
