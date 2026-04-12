import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../main.dart' show kBgPrimary, kTextMuted;
import '../providers/camera_provider.dart';

class CameraPreviewWidget extends ConsumerWidget {
  const CameraPreviewWidget({
    super.key,
    this.overlay,
    this.fit = BoxFit.cover,
  });

  final Widget? overlay;
  final BoxFit fit;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cameraInit = ref.watch(cameraInitProvider);

    return cameraInit.when(
      loading: () => _buildLoading(),
      error: (error, _) => _buildError(error),
      data: (_) {
        final service = ref.watch(cameraServiceProvider);
        if (!service.isInitialized || service.controller == null) {
          return _buildLoading();
        }
        return _buildPreview(service);
      },
    );
  }

  Widget _buildLoading() {
    return ColoredBox(
      color: kBgPrimary,
      child: const Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            CircularProgressIndicator(strokeWidth: 2),
            SizedBox(height: 16),
            Text(
              'Inizializzazione camera...',
              style: TextStyle(color: kTextMuted, fontSize: 14),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildError(Object error) {
    final message = error.toString();
    final truncated =
        message.length > 80 ? '${message.substring(0, 80)}…' : message;

    return ColoredBox(
      color: kBgPrimary,
      child: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.camera_outlined, color: kTextMuted, size: 48),
              const SizedBox(height: 16),
              Text(
                truncated,
                textAlign: TextAlign.center,
                style: const TextStyle(color: kTextMuted, fontSize: 13),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildPreview(dynamic service) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final preview = SizedBox(
          width: constraints.maxWidth,
          height: constraints.maxHeight,
          child: FittedBox(
            fit: fit,
            clipBehavior: Clip.hardEdge,
            child: SizedBox(
              width: service.controller!.value.previewSize?.height ?? constraints.maxWidth,
              height: service.controller!.value.previewSize?.width ?? constraints.maxHeight,
              child: CameraPreview(service.controller!),
            ),
          ),
        );

        if (overlay == null) {
          return preview;
        }

        return Stack(
          fit: StackFit.expand,
          children: [
            preview,
            overlay!,
          ],
        );
      },
    );
  }
}
