// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

// Riverpod provider for the equipment/device profile selected by the user.
// EquipmentProfile represents the first supported vertical: 3D printers.
// The schema is defined in assets/data/printer_profiles.json.
// Future verticals (automotive, HVAC, …) will extend this model or add
// domain-specific providers alongside this one.
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Describes the hardware profile of the device the user is working on.
///
/// Currently covers 3D printers; future verticals (automotive, HVAC, …)
/// will extend this model or introduce parallel domain-specific types.
class EquipmentProfile {
  final String id;
  final String name;
  final String manufacturer;
  final String extruderType;
  final bool enclosed;

  const EquipmentProfile({
    required this.id,
    required this.name,
    required this.manufacturer,
    required this.extruderType,
    required this.enclosed,
  });

  /// Parses a profile from a JSON map (e.g. from `assets/data/printer_profiles.json`).
  factory EquipmentProfile.fromJson(Map<String, dynamic> json) {
    return EquipmentProfile(
      id: json['id'] as String,
      name: json['name'] as String,
      manufacturer: json['manufacturer'] as String,
      extruderType: json['extruder_type'] as String? ?? '',
      enclosed: json['enclosed'] as bool? ?? false,
    );
  }

  /// Returns a copy with the given fields replaced.
  EquipmentProfile copyWith({
    String? id,
    String? name,
    String? manufacturer,
    String? extruderType,
    bool? enclosed,
  }) {
    return EquipmentProfile(
      id: id ?? this.id,
      name: name ?? this.name,
      manufacturer: manufacturer ?? this.manufacturer,
      extruderType: extruderType ?? this.extruderType,
      enclosed: enclosed ?? this.enclosed,
    );
  }
}

/// Manages the currently selected [EquipmentProfile].
///
/// State is `null` when no device has been selected yet.
class EquipmentNotifier extends StateNotifier<EquipmentProfile?> {
  EquipmentNotifier() : super(null);

  /// Sets [profile] as the active equipment profile.
  void select(EquipmentProfile profile) => state = profile;

  /// Creates and selects a custom profile with the given [name].
  void selectCustom(String name) => state = EquipmentProfile(
        id: 'custom',
        name: name,
        manufacturer: 'Custom',
        extruderType: '',
        enclosed: false,
      );

  /// Clears the active profile, returning state to `null`.
  void clear() => state = null;
}

/// Provides the [EquipmentNotifier] and current [EquipmentProfile] selection.
///
/// State is `null` until the user selects a device.
final equipmentProvider =
    StateNotifierProvider<EquipmentNotifier, EquipmentProfile?>(
  (ref) => EquipmentNotifier(),
);
