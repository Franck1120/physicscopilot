// Riverpod provider for the equipment/device profile selected by the user.
// The EquipmentProfile model is kept compatible with the 3D-printer JSON
// schema in assets/data/printer_profiles.json (an optional KB module).
import 'package:flutter_riverpod/flutter_riverpod.dart';

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

  factory EquipmentProfile.fromJson(Map<String, dynamic> json) {
    return EquipmentProfile(
      id: json['id'] as String,
      name: json['name'] as String,
      manufacturer: json['manufacturer'] as String,
      extruderType: json['extruder_type'] as String? ?? '',
      enclosed: json['enclosed'] as bool? ?? false,
    );
  }

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

class EquipmentNotifier extends StateNotifier<EquipmentProfile?> {
  EquipmentNotifier() : super(null);

  void select(EquipmentProfile profile) => state = profile;

  void selectCustom(String name) => state = EquipmentProfile(
        id: 'custom',
        name: name,
        manufacturer: 'Custom',
        extruderType: '',
        enclosed: false,
      );

  void clear() => state = null;
}

final equipmentProvider =
    StateNotifierProvider<EquipmentNotifier, EquipmentProfile?>(
  (ref) => EquipmentNotifier(),
);
