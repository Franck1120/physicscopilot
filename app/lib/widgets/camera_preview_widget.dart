import 'dart:typed_data';

import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../main.dart' show kBgPrimary, kTextMuted;
import '../providers/camera_provider.dart';

/// Full-screen camera feed with optional overlay and frame capture callback.
///
/// Watches [cameraInitProvider] to drive loading / error / live-preview states.
/// The underlying [CameraController] is accessed via [cameraServiceProvider]
/// (read-only — no rebuild on controller changes) to avoid unnecessary redraws.
class CameraPreviewWidget extends ConsumerWidget {
  /// Optional widget rendered on top of the camera feed (AR overlays, controls).
  final Widget? child;

  /// How the camera preview is fitted inside its parent.
  ///
  /// Defaults to [BoxFit.cover] so the feed fills the available space.
  final BoxFit fit;

  /// Called with the latest compressed frame bytes whenever the camera is ready.
  ///
  /// The callback fires only once per build when the controller is initialised.
  /// Use a [StreamSubscription] on [CameraService.frames] for continuous frames.
  final void Function(Uint8List bytes)? onFrameCaptured;

  const CameraPreviewWidget({
    super.key,
    this.child,
    this.fit = BoxFit.cover,
    this.onFrameCaptured,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final initAsync = ref.watch(cameraInitProvider);

    return initAsync.when(
      loading: _buildLoading,
      error: (_, __) => _buildError(),
      data: (_) => _buildPreview(ref),
    );
  }

  // ---------------------------------------------------------------------------
  // Loading state — spinner centred on dark background
  // ---------------------------------------------------------------------------

  Widget _buildLoading() {
    return const ColoredBox(
      color: kBgPrimary,
      child: Center(
        child: CircularProgressIndicator(),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Error state — camera icon + message on dark background
  // ---------------------------------------------------------------------------

  Widget _buildError() {
    return const ColoredBox(
      color: kBgPrimary,
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.camera_alt, color: kTextMuted, size: 48),
            SizedBox(height: 12),
            Text(
              'Camera non disponibile',
              style: TextStyle(color: kTextMuted, fontSize: 14),
            ),
          ],
        ),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // Live preview — AspectRatio + ClipRect + optional overlay Stack
  // ---------------------------------------------------------------------------

  Widget _buildPreview(WidgetRef ref) {
    // Read (not watch) to avoid rebuilding whenever internal controller state
    // changes (e.g. focus, exposure) — the init state is already tracked above.
    final controller = ref.read(cameraServiceProvider).controller;

    // Guard: controller should always be non-null after a successful init, but
    // fall back gracefully to avoid a null-dereference crash.
    if (controller == null || !controller.value.isInitialized) {
      return _buildError();
    }

    final preview = ClipRect(
      child: AspectRatio(
        aspectRatio: controller.value.aspectRatio,
        child: CameraPreview(controller),
      ),
    );

    // Wrap in a Stack only when an overlay child is provided to keep the
    // widget tree shallow for the common no-overlay case.
    if (child == null) return preview;

    return ClipRect(
      child: AspectRatio(
        aspectRatio: controller.value.aspectRatio,
        child: Stack(
          fit: StackFit.expand,
          children: [
            CameraPreview(controller),
            child!,
          ],
        ),
      ),
    );
  }
}
