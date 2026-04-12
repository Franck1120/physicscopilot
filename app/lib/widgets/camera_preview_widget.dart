import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../main.dart' show kAccent, kTextMuted;
import '../providers/camera_provider.dart';

class CameraPreviewWidget extends ConsumerWidget {
  const CameraPreviewWidget({
    super.key,
    this.overlay,
    this.onCapture,
    this.borderRadius,
    this.fit,
  });

  /// Optional widget rendered above the camera feed (e.g. AR guide).
  final Widget? overlay;

  /// When non-null, a capture FAB is shown at bottom-right.
  final VoidCallback? onCapture;

  /// When non-null, the whole widget is clipped to this radius.
  final BorderRadius? borderRadius;

  /// How to inscribe the camera feed into the available space.
  /// Defaults to [BoxFit.cover].
  final BoxFit? fit;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cameraInit = ref.watch(cameraInitProvider);
    final cameraService = ref.watch(cameraServiceProvider);

    final content = cameraInit.when(
      loading: () => const _CameraLoadingState(),
      error: (error, _) => _CameraErrorState(message: error.toString()),
      data: (_) {
        final controller = cameraService.controller;
        if (controller == null || !controller.value.isInitialized) {
          return const _CameraLoadingState();
        }
        return _CameraPreview(
          controller: controller,
          overlay: overlay,
          onCapture: onCapture,
          fit: fit,
        );
      },
    );

    if (borderRadius != null) {
      return ClipRRect(borderRadius: borderRadius!, child: content);
    }
    return content;
  }
}

// ---------------------------------------------------------------------------
// Internal preview with stack layers
// ---------------------------------------------------------------------------

class _CameraPreview extends StatelessWidget {
  const _CameraPreview({
    required this.controller,
    required this.overlay,
    required this.onCapture,
    required this.fit,
  });

  final CameraController controller;
  final Widget? overlay;
  final VoidCallback? onCapture;
  final BoxFit? fit;

  @override
  Widget build(BuildContext context) {
    return Stack(
      fit: StackFit.expand,
      children: [
        // Layer 0 — full-coverage camera feed
        ClipRect(
          child: OverflowBox(
            maxWidth: double.infinity,
            maxHeight: double.infinity,
            child: FittedBox(
              fit: fit ?? BoxFit.cover,
              child: SizedBox(
                width: controller.value.previewSize?.height ?? 1,
                height: controller.value.previewSize?.width ?? 1,
                child: CameraPreview(controller),
              ),
            ),
          ),
        ),

        // Layer 1 — optional overlay (e.g. AR guide)
        if (overlay != null) Positioned.fill(child: overlay!),

        // Layer 2 — optional capture FAB
        if (onCapture != null)
          Positioned(
            bottom: 16,
            right: 16,
            child: FloatingActionButton(
              heroTag: 'camera_preview_capture',
              backgroundColor: kAccent,
              foregroundColor: Colors.white,
              onPressed: onCapture,
              child: const Icon(Icons.camera_alt),
            ),
          ),
      ],
    );
  }
}

// ---------------------------------------------------------------------------
// Placeholder states
// ---------------------------------------------------------------------------

class _CameraLoadingState extends StatelessWidget {
  const _CameraLoadingState();

  @override
  Widget build(BuildContext context) {
    return Container(
      color: const Color(0xFF0D0D0D),
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CircularProgressIndicator(color: kAccent),
            const SizedBox(height: 16),
            Text(
              'Inizializzazione camera…',
              style: TextStyle(color: kTextMuted, fontSize: 13),
            ),
          ],
        ),
      ),
    );
  }
}

class _CameraErrorState extends StatelessWidget {
  const _CameraErrorState({this.message});

  final String? message;

  @override
  Widget build(BuildContext context) {
    return Container(
      color: const Color(0xFF0D0D0D),
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.camera_alt_outlined, color: kTextMuted, size: 48),
            const SizedBox(height: 12),
            Text(
              'Camera non disponibile',
              style: TextStyle(color: kTextMuted, fontSize: 13),
            ),
            if (message != null) ...[
              const SizedBox(height: 6),
              Text(
                message!,
                style: TextStyle(color: kTextMuted, fontSize: 11),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                textAlign: TextAlign.center,
              ),
            ],
          ],
        ),
      ),
    );
  }
}
