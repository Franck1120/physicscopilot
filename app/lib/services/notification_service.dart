// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

import 'package:flutter_local_notifications/flutter_local_notifications.dart';

/// Thin wrapper around [FlutterLocalNotificationsPlugin].
///
/// Call [initialize] once at app start (inside [main]) before showing
/// any notifications.  All public methods are safe to call from any
/// async context — errors are silently swallowed so a notification
/// failure never crashes the app.
class NotificationService {
  NotificationService._();

  static final _plugin = FlutterLocalNotificationsPlugin();
  static bool _ready = false;

  // ── Android notification channel ─────────────────────────────────────────

  static const _channelId = 'session_channel';
  static const _channelName = 'Sessione attiva';
  static const _notifId = 1;

  // ── Initialisation ───────────────────────────────────────────────────────

  /// Initialise the plugin. Safe to call multiple times — runs only once.
  static Future<void> initialize() async {
    if (_ready) return;
    const android = AndroidInitializationSettings('@mipmap/ic_launcher');
    const settings = InitializationSettings(android: android);
    await _plugin.initialize(settings);
    _ready = true;
  }

  // ── Public API ───────────────────────────────────────────────────────────

  /// Shows a persistent "session in progress" notification.
  ///
  /// The notification is non-dismissable ([ongoing]) so the user always
  /// has a path back to the app while a session is running in the background.
  static Future<void> showSessionRunning() async {
    if (!_ready) return;
    try {
      const androidDetails = AndroidNotificationDetails(
        _channelId,
        _channelName,
        channelDescription: 'PhysicsCopilot ha una sessione aperta in background',
        importance: Importance.low,
        priority: Priority.low,
        ongoing: true,
        autoCancel: false,
        showWhen: false,
      );
      await _plugin.show(
        _notifId,
        'PhysicsCopilot',
        'Sessione in corso — tocca per tornare all\'analisi',
        const NotificationDetails(android: androidDetails),
      );
    } catch (_) {}
  }

  /// Dismisses the session notification (call on foreground resume or end).
  static Future<void> cancelSessionNotification() async {
    if (!_ready) return;
    try {
      await _plugin.cancel(_notifId);
    } catch (_) {}
  }
}
