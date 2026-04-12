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

    // ── Brand / manufacturer filtering (via select pattern) ────────────────

    test('select() profile with manufacturer HP — state has matching manufacturer',
        () {
      const hp = EquipmentProfile(
        id: 'hp_laserjet',
        name: 'HP LaserJet Pro',
        manufacturer: 'HP',
        extruderType: '',
        enclosed: false,
      );

      notifier.select(hp);
      final state = container.read(equipmentProvider);

      expect(state?.manufacturer, equals('HP'));
    });

    test('selecting one brand replaces a previous brand selection', () {
      const hp = EquipmentProfile(
        id: 'hp_laserjet',
        name: 'HP LaserJet Pro',
        manufacturer: 'HP',
        extruderType: '',
        enclosed: false,
      );
      const epson = EquipmentProfile(
        id: 'epson_wf',
        name: 'Epson WorkForce',
        manufacturer: 'Epson',
        extruderType: '',
        enclosed: false,
      );

      notifier.select(hp);
      expect(container.read(equipmentProvider)?.manufacturer, 'HP');

      notifier.select(epson);
      expect(container.read(equipmentProvider)?.manufacturer, 'Epson');
    });

    test('clear() after brand selection resets to null (filter reset)', () {
      const hp = EquipmentProfile(
        id: 'hp_laserjet',
        name: 'HP LaserJet Pro',
        manufacturer: 'HP',
        extruderType: '',
        enclosed: false,
      );

      notifier.select(hp);
      expect(container.read(equipmentProvider), isNotNull);

      notifier.clear();
      expect(container.read(equipmentProvider), isNull);
    });

    // ── selectedDevice getter (state access) ───────────────────────────────

    test('state property exposes the selected profile (selectedDevice pattern)',
        () {
      const profile = EquipmentProfile(
        id: 'bambu_p1s',
        name: 'Bambu P1S',
        manufacturer: 'Bambu Lab',
        extruderType: 'direct',
        enclosed: true,
      );

      notifier.select(profile);

      // The notifier's state == container.read() — acts as selectedDevice getter.
      final selected = container.read(equipmentProvider);
      expect(selected?.id, 'bambu_p1s');
      expect(selected?.name, 'Bambu P1S');
    });

    test('state is null before any selection — no selectedDevice', () {
      expect(container.read(equipmentProvider), isNull);
    });

    // ── Selecting a non-existent / unknown ID stays stable ─────────────────

    test('selecting any profile never crashes — unknown manufacturer stays stable',
        () {
      const unknownProfile = EquipmentProfile(
        id: 'unknown_xyz',
        name: 'Unknown Device',
        manufacturer: 'UnknownCo',
        extruderType: 'unknown',
        enclosed: false,
      );

      // Should not throw regardless of how unfamiliar the profile is.
      expect(() => notifier.select(unknownProfile), returnsNormally);
      final state = container.read(equipmentProvider);
      expect(state?.id, 'unknown_xyz');
    });

    test('selecting same profile twice leaves state identical', () {
      const profile = EquipmentProfile(
        id: 'mk4',
        name: 'MK4',
        manufacturer: 'Prusa',
        extruderType: 'direct',
        enclosed: false,
      );

      notifier.select(profile);
      notifier.select(profile);

      final state = container.read(equipmentProvider);
      expect(state?.id, 'mk4');
    });

    test('selectCustom() then clear() returns to null — no crash', () {
      notifier.selectCustom('Ender 5 Plus DIY');
      expect(container.read(equipmentProvider)?.id, 'custom');

      notifier.clear();
      expect(container.read(equipmentProvider), isNull);
    });
  });
}
