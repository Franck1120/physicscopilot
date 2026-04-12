// Copyright (c) 2026 Franck1120. All rights reserved.
// Use of this source code is governed by a MIT license that can be
// found in the LICENSE file.

/// Represents a physical device (e.g. a printer, sensor, or appliance) owned by a user.
///
/// Maps directly to the `devices` table in Supabase.
class Device {
  final String id;
  final String userId;
  final String brand;
  final String model;
  final DateTime createdAt;

  const Device({
    required this.id,
    required this.userId,
    required this.brand,
    required this.model,
    required this.createdAt,
  });

  // ---------------------------------------------------------------------------
  // Serialization
  // ---------------------------------------------------------------------------

  factory Device.fromJson(Map<String, dynamic> json) => Device(
        id: json['id'] as String,
        userId: json['user_id'] as String,
        brand: json['brand'] as String,
        model: json['model'] as String,
        createdAt: DateTime.parse(json['created_at'] as String),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'user_id': userId,
        'brand': brand,
        'model': model,
        'created_at': createdAt.toIso8601String(),
      };

  // ---------------------------------------------------------------------------
  // Convenience
  // ---------------------------------------------------------------------------

  /// Human-readable label shown in the UI: "Brand Model".
  String get displayName => '$brand $model';

  // ---------------------------------------------------------------------------
  // copyWith
  // ---------------------------------------------------------------------------

  Device copyWith({
    String? id,
    String? userId,
    String? brand,
    String? model,
    DateTime? createdAt,
  }) =>
      Device(
        id: id ?? this.id,
        userId: userId ?? this.userId,
        brand: brand ?? this.brand,
        model: model ?? this.model,
        createdAt: createdAt ?? this.createdAt,
      );

  // ---------------------------------------------------------------------------
  // Equality
  // ---------------------------------------------------------------------------

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is Device &&
          runtimeType == other.runtimeType &&
          id == other.id &&
          userId == other.userId &&
          brand == other.brand &&
          model == other.model &&
          createdAt == other.createdAt;

  @override
  int get hashCode =>
      Object.hash(id, userId, brand, model, createdAt);

  @override
  String toString() =>
      'Device(id: $id, userId: $userId, brand: $brand, model: $model, '
      'createdAt: $createdAt)';
}
