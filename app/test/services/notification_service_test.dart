// Unit tests for NotificationService.
//
// flutter_local_notifications requires Android/iOS platform channels.
// These tests verify only the guard logic (_ready flag) that prevents crashes
// when the plugin is not initialised, without invoking native code.
import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/services/notification_service.dart';

void main() {
  setUpAll(() {
    TestWidgetsFlutterBinding.ensureInitialized();
  });

  group('NotificationService — guard logic', () {
    test('showSessionRunning() before initialize() does not throw', () async {
      // _ready == false → method returns early without calling the plugin.
      await expectLater(NotificationService.showSessionRunning(), completes);
    });

    test('cancelSessionNotification() before initialize() does not throw',
        () async {
      await expectLater(
          NotificationService.cancelSessionNotification(), completes,);
    });

    test('showSessionRunning() is callable as a static method', () {
      // Verify the static surface is correct.
      expect(NotificationService.showSessionRunning, isA<Function>());
    });

    test('cancelSessionNotification() is callable as a static method', () {
      expect(NotificationService.cancelSessionNotification, isA<Function>());
    });

    test('initialize() is callable as a static method', () {
      expect(NotificationService.initialize, isA<Function>());
    });

    test('multiple calls to showSessionRunning() before init do not throw',
        () async {
      await NotificationService.showSessionRunning();
      await expectLater(NotificationService.showSessionRunning(), completes);
    });

    test(
        'multiple calls to cancelSessionNotification() before init do not throw',
        () async {
      await NotificationService.cancelSessionNotification();
      await expectLater(
          NotificationService.cancelSessionNotification(), completes,);
    });
  });
}
