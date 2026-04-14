import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:physicscopilot/widgets/step_progress.dart';

// ── State ─────────────────────────────────────────────────────────────────────

/// Immutable state for the multi-step guided procedure.
class ProcedureState {
  final List<StepInfo> steps;

  /// Zero-based index of the step currently being executed.
  final int currentIndex;

  const ProcedureState({
    this.steps = const [],
    this.currentIndex = 0,
  });

  /// Returns a copy with the given fields replaced.
  ProcedureState copyWith({
    List<StepInfo>? steps,
    int? currentIndex,
  }) =>
      ProcedureState(
        steps: steps ?? this.steps,
        currentIndex: currentIndex ?? this.currentIndex,
      );

  /// Total number of steps in the current procedure.
  int get totalSteps => steps.length;

  /// True when [currentIndex] has reached the last step.
  bool get isCompleted =>
      steps.isNotEmpty && currentIndex >= steps.length - 1;
}

// ── Notifier ──────────────────────────────────────────────────────────────────

/// Drives step-by-step procedure navigation.
///
/// Steps are populated from server `response` payloads via [updateFromResponse]
/// or set directly via [loadSteps]. Navigation is done with [advance] / [goTo].
class StepNotifier extends StateNotifier<ProcedureState> {
  StepNotifier() : super(const ProcedureState());

  /// Replaces the current procedure with [steps] and resets to step 0.
  void loadSteps(List<StepInfo> steps) {
    state = ProcedureState(steps: steps, currentIndex: 0);
  }

  /// Moves to the next step. No-op when already on the last step.
  void advance() {
    if (state.currentIndex < state.steps.length - 1) {
      state = state.copyWith(currentIndex: state.currentIndex + 1);
    }
  }

  /// Jumps to [index]. No-op when out of bounds.
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

  /// Resets to an empty procedure state.
  void reset() => state = const ProcedureState();
}

// ── Provider ──────────────────────────────────────────────────────────────────

/// Provides the [StepNotifier] and current [ProcedureState].
///
/// Updated by the session screen when the server returns a multi-step response.
final stepProvider =
    StateNotifierProvider<StepNotifier, ProcedureState>((_) => StepNotifier());
