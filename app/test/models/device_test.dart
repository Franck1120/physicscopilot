import 'package:flutter_test/flutter_test.dart';
import 'package:physicscopilot/models/device.dart';

void main() {
  final _createdAt = DateTime(2026, 1, 15, 10, 30, 0);

  final _deviceJson = <String, dynamic>{
    'id': 'device-001',
    'user_id': 'user-abc',
    'brand': 'HP',
    'model': 'LaserJet 1020',
    'created_at': '2026-01-15T10:30:00.000',
  };

  Device _makeDevice({
    String id = 'device-001',
    String userId = 'user-abc',
    String brand = 'HP',
    String model = 'LaserJet 1020',
    DateTime? createdAt,
  }) =>
      Device(
        id: id,
        userId: userId,
        brand: brand,
        model: model,
        createdAt: createdAt ?? DateTime(2026, 1, 15, 10, 30, 0),
      );

  group('Device.fromJson', () {
    test('parses all fields correctly', () {
      final device = Device.fromJson(_deviceJson);

      expect(device.id, 'device-001');
      expect(device.userId, 'user-abc');
      expect(device.brand, 'HP');
      expect(device.model, 'LaserJet 1020');
      expect(device.createdAt, DateTime.parse('2026-01-15T10:30:00.000'));
    });

    test('parses device_id with underscored key (user_id)', () {
      final json = Map<String, dynamic>.from(_deviceJson)
        ..['user_id'] = 'user-xyz';
      final device = Device.fromJson(json);
      expect(device.userId, 'user-xyz');
    });
  });

  group('Device.toJson', () {
    test('serializes all fields to map', () {
      final device = _makeDevice(createdAt: _createdAt);
      final json = device.toJson();

      expect(json['id'], 'device-001');
      expect(json['user_id'], 'user-abc');
      expect(json['brand'], 'HP');
      expect(json['model'], 'LaserJet 1020');
      expect(json['created_at'], _createdAt.toIso8601String());
    });

    test('round-trips through fromJson → toJson', () {
      final original = Device.fromJson(_deviceJson);
      final roundTripped = Device.fromJson(original.toJson());
      expect(roundTripped, original);
    });
  });

  group('Device.copyWith', () {
    test('returns identical device when no overrides provided', () {
      final device = _makeDevice();
      final copy = device.copyWith();
      expect(copy, device);
    });

    test('overrides only the specified fields', () {
      final device = _makeDevice();
      final copy = device.copyWith(brand: 'Canon', model: 'PIXMA');

      expect(copy.brand, 'Canon');
      expect(copy.model, 'PIXMA');
      expect(copy.id, device.id);
      expect(copy.userId, device.userId);
      expect(copy.createdAt, device.createdAt);
    });

    test('overrides id', () {
      final device = _makeDevice();
      final copy = device.copyWith(id: 'new-id');
      expect(copy.id, 'new-id');
      expect(copy.brand, device.brand);
    });

    test('overrides userId', () {
      final device = _makeDevice();
      final copy = device.copyWith(userId: 'new-user');
      expect(copy.userId, 'new-user');
    });

    test('overrides createdAt', () {
      final device = _makeDevice();
      final newDate = DateTime(2025, 6, 1);
      final copy = device.copyWith(createdAt: newDate);
      expect(copy.createdAt, newDate);
    });
  });

  group('Device equality', () {
    test('two devices with identical fields are equal', () {
      final a = _makeDevice();
      final b = _makeDevice();
      expect(a, equals(b));
    });

    test('devices with different id are not equal', () {
      final a = _makeDevice(id: 'id-1');
      final b = _makeDevice(id: 'id-2');
      expect(a, isNot(equals(b)));
    });

    test('devices with different brand are not equal', () {
      final a = _makeDevice(brand: 'HP');
      final b = _makeDevice(brand: 'Canon');
      expect(a, isNot(equals(b)));
    });

    test('equal devices have equal hash codes', () {
      final a = _makeDevice();
      final b = _makeDevice();
      expect(a.hashCode, b.hashCode);
    });

    test('device is equal to itself (identity)', () {
      final device = _makeDevice();
      expect(device, equals(device));
    });
  });

  group('Device.displayName', () {
    test('concatenates brand and model with a space', () {
      final device = _makeDevice(brand: 'HP', model: 'LaserJet 1020');
      expect(device.displayName, 'HP LaserJet 1020');
    });

    test('handles single-word brand and model', () {
      final device = _makeDevice(brand: 'Canon', model: 'PIXMA');
      expect(device.displayName, 'Canon PIXMA');
    });
  });

  group('Device.toString', () {
    test('includes all field values', () {
      final device = _makeDevice();
      final str = device.toString();

      expect(str, contains('device-001'));
      expect(str, contains('user-abc'));
      expect(str, contains('HP'));
      expect(str, contains('LaserJet 1020'));
    });

    test('starts with Device(', () {
      final device = _makeDevice();
      expect(device.toString(), startsWith('Device('));
    });
  });
}
