import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:physicscopilot/providers/equipment_provider.dart';

void main() {
  group('EquipmentProfile', () {
    test('fromJson parses all fields correctly', () {
      final profile = EquipmentProfile.fromJson({
        'id': 'prusa_mk4',
        'name': 'Prusa MK4',
        'manufacturer': 'Prusa Research',
        'extruder_type': 'direct',
        'enclosed': true,
      });

      expect(profile.id, equals('prusa_mk4'));
      expect(profile.name, equals('Prusa MK4'));
      expect(profile.manufacturer, equals('Prusa Research'));
      expect(profile.extruderType, equals('direct'));
      expect(profile.enclosed, isTrue);
    });

    test('fromJson uses defaults for missing optional fields', () {
      final profile = EquipmentProfile.fromJson({
        'id': 'test',
        'name': 'Test Printer',
        'manufacturer': 'Acme',
      });

      expect(profile.extruderType, equals(''));
      expect(profile.enclosed, isFalse);
    });

    test('copyWith replaces only the specified fields', () {
      const original = EquipmentProfile(
        id: 'mk4',
        name: 'MK4',
        manufacturer: 'Prusa',
        extruderType: 'direct',
        enclosed: false,
      );

      final copy = original.copyWith(enclosed: true, name: 'MK4S');

      expect(copy.id, equals('mk4'));
      expect(copy.name, equals('MK4S'));
      expect(copy.manufacturer, equals('Prusa'));
      expect(copy.extruderType, equals('direct'));
      expect(copy.enclosed, isTrue);
    });
  });

  group('EquipmentNotifier', () {
    late ProviderContainer container;
    late EquipmentNotifier notifier;

    setUp(() {
      container = ProviderContainer();
      notifier = container.read(equipmentProvider.notifier);
    });

    tearDown(() => container.dispose());

    test('initial state is null', () {
      expect(container.read(equipmentProvider), isNull);
    });

    test('select() sets the active profile', () {
      const profile = EquipmentProfile(
        id: 'mk4',
        name: 'Prusa MK4',
        manufacturer: 'Prusa Research',
        extruderType: 'direct',
        enclosed: false,
      );

      notifier.select(profile);

      expect(container.read(equipmentProvider), equals(profile));
    });

    test('select() replaces a previously selected profile', () {
      const first = EquipmentProfile(
        id: 'mk4',
        name: 'MK4',
        manufacturer: 'Prusa',
        extruderType: 'direct',
        enclosed: false,
      );
      const second = EquipmentProfile(
        id: 'bambu_x1',
        name: 'X1 Carbon',
        manufacturer: 'Bambu Lab',
        extruderType: 'direct',
        enclosed: true,
      );

      notifier.select(first);
      notifier.select(second);

      final state = container.read(equipmentProvider);
      expect(state?.id, equals('bambu_x1'));
      expect(state?.enclosed, isTrue);
    });

    test('selectCustom() creates a profile with id="custom" and supplied name', () {
      notifier.selectCustom('My DIY Printer');

      final state = container.read(equipmentProvider);
      expect(state?.id, equals('custom'));
      expect(state?.name, equals('My DIY Printer'));
      expect(state?.manufacturer, equals('Custom'));
      expect(state?.extruderType, equals(''));
      expect(state?.enclosed, isFalse);
    });

    test('clear() resets state to null', () {
      const profile = EquipmentProfile(
        id: 'mk4',
        name: 'MK4',
        manufacturer: 'Prusa',
        extruderType: 'direct',
        enclosed: false,
      );
      notifier.select(profile);
      expect(container.read(equipmentProvider), isNotNull);

      notifier.clear();

      expect(container.read(equipmentProvider), isNull);
    });

    test('clear() when already null does not throw', () {
      expect(() => notifier.clear(), returnsNormally);
      expect(container.read(equipmentProvider), isNull);
    });
  });
}
