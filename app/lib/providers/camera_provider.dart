// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../services/camera_service.dart' show CameraService, FrameQuality;

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

/// Live stream of the current [FrameQuality] assessed by [CameraService].
final frameQualityProvider = StreamProvider<FrameQuality>((ref) {
  return ref.watch(cameraServiceProvider).quality;
});
