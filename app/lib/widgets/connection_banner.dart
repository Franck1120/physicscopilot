// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'package:flutter/material.dart';

import '../services/websocket_service.dart' show ConnectionStatus;

/// Thin banner shown at the top of the screen when the WebSocket is not connected.
///
/// Displays an orange indicator while [ConnectionStatus.connecting] and a red
/// indicator when [ConnectionStatus.disconnected]. The widget is a no-op at
/// [ConnectionStatus.connected] — callers should hide it entirely in that case.
class ConnectionBanner extends StatelessWidget {
  const ConnectionBanner({super.key, required this.status});
  final ConnectionStatus status;

  @override
  Widget build(BuildContext context) {
    final isConnecting = status == ConnectionStatus.connecting;
    final color = isConnecting ? Colors.orangeAccent : Colors.redAccent;
    final icon = isConnecting ? Icons.wifi_find : Icons.wifi_off;
    final message = isConnecting
        ? 'Connessione al server in corso…'
        : 'Server non raggiungibile — i dati non vengono inviati';

    return Semantics(
      liveRegion: true,
      label: message,
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 7),
        color: color.withAlpha(30),
        child: ExcludeSemantics(
          child: Row(
            children: [
              Icon(icon, color: color, size: 14),
              const SizedBox(width: 8),
              Expanded(
                child: Text(message,
                    style: TextStyle(
                        color: color, fontSize: 12, fontWeight: FontWeight.w500)),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
