import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../services/camera_service.dart';

/// Singleton [CameraService]; disposed with the provider scope.
final cameraServiceProvider = Provider<CameraService>((ref) {
  final service = CameraService();
  ref.onDispose(() => service.dispose());
  return service;
});

/// Drives camera initialisation; resolves to [AsyncValue.data] when ready.
///
/// Downstream widgets watch this to know when the camera is usable.
final cameraInitProvider = FutureProvider<void>((ref) async {
  final service = ref.watch(cameraServiceProvider);
  await service.initialize();
});
