import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../widgets/step_progress.dart';

// ── State ─────────────────────────────────────────────────────────────────────

class ProcedureState {
  final List<StepInfo> steps;
  final int currentIndex;

  const ProcedureState({
    this.steps = const [],
    this.currentIndex = 0,
  });

  ProcedureState copyWith({
    List<StepInfo>? steps,
    int? currentIndex,
  }) =>
      ProcedureState(
        steps: steps ?? this.steps,
        currentIndex: currentIndex ?? this.currentIndex,
      );

  int get totalSteps => steps.length;
  bool get isCompleted =>
      steps.isNotEmpty && currentIndex >= steps.length - 1;
}

// ── Notifier ──────────────────────────────────────────────────────────────────

class StepNotifier extends StateNotifier<ProcedureState> {
  StepNotifier() : super(const ProcedureState());

  void loadSteps(List<StepInfo> steps) {
    state = ProcedureState(steps: steps, currentIndex: 0);
  }

  void advance() {
    if (state.currentIndex < state.steps.length - 1) {
      state = state.copyWith(currentIndex: state.currentIndex + 1);
    }
  }

  void goTo(int index) {
    if (index >= 0 && index < state.steps.length) {
      state = state.copyWith(currentIndex: index);
    }
  }

  /// Applies step data from a server `response` payload.
  ///
  /// Expected shape:
  /// ```json
  /// {
  ///   "steps": [{"description": "...", "estimated_seconds": 30}],
  ///   "current_step": 0
  /// }
  /// ```
  void updateFromResponse(Map<String, dynamic> json) {
    final stepsJson = json['steps'] as List<dynamic>?;
    if (stepsJson != null) {
      final steps = stepsJson.map((s) {
        final map = s as Map<String, dynamic>;
        final seconds = map['estimated_seconds'] as int?;
        return StepInfo(
          description: (map['description'] as String?) ?? '',
          estimatedDuration:
              seconds != null ? Duration(seconds: seconds) : null,
        );
      }).toList();
      loadSteps(steps);
    }

    final currentStep = json['current_step'] as int?;
    if (currentStep != null) goTo(currentStep);
  }

  void reset() => state = const ProcedureState();
}

// ── Provider ──────────────────────────────────────────────────────────────────

final stepProvider =
    StateNotifierProvider<StepNotifier, ProcedureState>((_) => StepNotifier());
