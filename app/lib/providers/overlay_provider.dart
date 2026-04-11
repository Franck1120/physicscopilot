import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/session_provider.dart';
import '../widgets/ar_overlay.dart';

/// Derives the current [OverlayData] from the session's raw overlay map.
/// Returns `null` when no overlay is active.
final overlayDataProvider = Provider<OverlayData?>((ref) {
  final overlayMap = ref.watch(sessionProvider).overlay;
  if (overlayMap == null) return null;
  return OverlayData.fromJson(overlayMap);
});
