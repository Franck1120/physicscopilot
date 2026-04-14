import 'package:flutter/material.dart';

import 'package:physicscopilot/main.dart' show kAccent, kBgCard, kBgCardBorder, kTextMuted;

/// Returns a list of step strings if [text] contains 2+ numbered items
/// (e.g. "1. Do this\n2. Do that"), otherwise null.
List<String>? parseSteps(String text) {
  final lines = text.split('\n');
  final steps = <String>[];
  final stepRe = RegExp(r'^\s*(\d+)[.)]\s+(.+)$');
  for (final line in lines) {
    final m = stepRe.firstMatch(line);
    if (m != null) steps.add(m.group(2)!.trim());
  }
  return steps.length >= 2 ? steps : null;
}

/// Interactive checklist widget for multi-step AI responses.
///
/// Renders each step as a tappable card; checked steps are visually struck
/// through and highlighted. Shows a completion banner when all steps are done.
class MultiStepView extends StatefulWidget {
  const MultiStepView({super.key, required this.steps});
  final List<String> steps;

  @override
  State<MultiStepView> createState() => _MultiStepViewState();
}

class _MultiStepViewState extends State<MultiStepView> {
  final Set<int> _checked = {};

  @override
  Widget build(BuildContext context) {
    final total = widget.steps.length;
    final done = _checked.length;
    final progress = total == 0 ? 0.0 : done / total;
    final allDone = done == total;

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          // Header + progress
          Row(
            children: [
              const Icon(Icons.format_list_numbered,
                  color: kAccent, size: 16,),
              const SizedBox(width: 8),
              Text(
                '$done / $total completati',
                style: const TextStyle(
                  color: kAccent,
                  fontSize: 11,
                  fontWeight: FontWeight.w600,
                  letterSpacing: 0.4,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: progress,
              minHeight: 4,
              backgroundColor: kBgCardBorder,
              valueColor: const AlwaysStoppedAnimation<Color>(kAccent),
            ),
          ),
          const SizedBox(height: 12),
          // Step cards
          ...List.generate(total, (i) {
            final isDone = _checked.contains(i);
            return Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: GestureDetector(
                onTap: () => setState(() {
                  if (isDone) {
                    _checked.remove(i);
                  } else {
                    _checked.add(i);
                  }
                }),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.symmetric(
                      horizontal: 12, vertical: 10,),
                  decoration: BoxDecoration(
                    color: isDone
                        ? kAccent.withAlpha(30)
                        : kBgCard,
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(
                      color: isDone ? kAccent.withAlpha(120) : kBgCardBorder,
                    ),
                  ),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      AnimatedContainer(
                        duration: const Duration(milliseconds: 200),
                        width: 22,
                        height: 22,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: isDone ? kAccent : Colors.transparent,
                          border: Border.all(
                            color: isDone ? kAccent : kTextMuted,
                            width: 1.5,
                          ),
                        ),
                        child: isDone
                            ? const Icon(Icons.check,
                                color: Colors.white, size: 13,)
                            : Center(
                                child: Text(
                                  '${i + 1}',
                                  style: const TextStyle(
                                      color: kTextMuted,
                                      fontSize: 11,
                                      fontWeight: FontWeight.w600,),
                                ),
                              ),
                      ),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          widget.steps[i],
                          style: TextStyle(
                            color: isDone
                                ? Colors.white.withAlpha(140)
                                : Colors.white,
                            fontSize: 13,
                            height: 1.5,
                            decoration: isDone
                                ? TextDecoration.lineThrough
                                : TextDecoration.none,
                            decorationColor: Colors.white38,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            );
          }),
          if (allDone) ...[
            const SizedBox(height: 8),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.symmetric(vertical: 10),
              decoration: BoxDecoration(
                color: kAccent.withAlpha(40),
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: kAccent.withAlpha(100)),
              ),
              child: const Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.check_circle_outline,
                      color: kAccent, size: 16,),
                  SizedBox(width: 8),
                  Text(
                    'Tutti i passi completati!',
                    style: TextStyle(
                        color: kAccent,
                        fontSize: 13,
                        fontWeight: FontWeight.w600,),
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }
}
