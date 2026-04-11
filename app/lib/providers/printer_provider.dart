// Provider Riverpod per la stampante selezionata dall'utente
import 'package:flutter_riverpod/flutter_riverpod.dart';

class PrinterProfile {
  final String id;
  final String name;
  final String manufacturer;
  final String extruderType;
  final bool enclosed;

  const PrinterProfile({
    required this.id,
    required this.name,
    required this.manufacturer,
    required this.extruderType,
    required this.enclosed,
  });

  factory PrinterProfile.fromJson(Map<String, dynamic> json) {
    return PrinterProfile(
      id: json['id'] as String,
      name: json['name'] as String,
      manufacturer: json['manufacturer'] as String,
      extruderType: json['extruder_type'] as String,
      enclosed: json['enclosed'] as bool,
    );
  }

  PrinterProfile copyWith({
    String? id,
    String? name,
    String? manufacturer,
    String? extruderType,
    bool? enclosed,
  }) {
    return PrinterProfile(
      id: id ?? this.id,
      name: name ?? this.name,
      manufacturer: manufacturer ?? this.manufacturer,
      extruderType: extruderType ?? this.extruderType,
      enclosed: enclosed ?? this.enclosed,
    );
  }
}

class PrinterNotifier extends StateNotifier<PrinterProfile?> {
  PrinterNotifier() : super(null);

  void select(PrinterProfile printer) => state = printer;

  void selectCustom(String name) => state = PrinterProfile(
        id: 'custom',
        name: name,
        manufacturer: 'Custom',
        extruderType: 'direct_drive',
        enclosed: false,
      );

  void clear() => state = null;
}

final printerProvider =
    StateNotifierProvider<PrinterNotifier, PrinterProfile?>(
  (ref) => PrinterNotifier(),
);
