import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/equipment_provider.dart';

const _kAccent = Color(0xFF1B4F72);
const _kBackground = Color(0xFF0D1B2A);
const _kSurface = Color(0xFF1A2B3C);
const _kOnSurface = Color(0xFFE0E8F0);

class EquipmentSelectionScreen extends ConsumerStatefulWidget {
  final VoidCallback onComplete;

  const EquipmentSelectionScreen({super.key, required this.onComplete});

  @override
  ConsumerState<EquipmentSelectionScreen> createState() =>
      _EquipmentSelectionScreenState();
}

class _EquipmentSelectionScreenState
    extends ConsumerState<EquipmentSelectionScreen> {
  List<EquipmentProfile> _allProfiles = [];
  List<EquipmentProfile> _filtered = [];
  final TextEditingController _searchController = TextEditingController();
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadProfiles();
    _searchController.addListener(_onSearchChanged);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadProfiles() async {
    final raw =
        await rootBundle.loadString('assets/data/printer_profiles.json');
    final decoded = jsonDecode(raw) as Map<String, dynamic>;
    final list = (decoded['profiles'] as List<dynamic>)
        .map((e) => EquipmentProfile.fromJson(e as Map<String, dynamic>))
        .toList();
    setState(() {
      _allProfiles = list;
      _filtered = list;
      _loading = false;
    });
  }

  void _onSearchChanged() {
    final query = _searchController.text.toLowerCase();
    setState(() {
      _filtered = _allProfiles
          .where((p) => p.name.toLowerCase().contains(query))
          .toList();
    });
  }

  void _selectEquipment(EquipmentProfile profile) {
    ref.read(equipmentProvider.notifier).select(profile);
    widget.onComplete();
  }

  Future<void> _showCustomDialog() async {
    final controller = TextEditingController();
    await showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: _kSurface,
        title: const Text(
          'Nome dispositivo',
          style: TextStyle(color: _kOnSurface),
        ),
        content: TextField(
          controller: controller,
          autofocus: true,
          style: const TextStyle(color: _kOnSurface),
          cursorColor: _kAccent,
          decoration: const InputDecoration(
            hintText: 'Es. Trapano Makita, Caldaia Vaillant…',
            hintStyle: TextStyle(color: Colors.white38),
            enabledBorder: UnderlineInputBorder(
              borderSide: BorderSide(color: _kAccent),
            ),
            focusedBorder: UnderlineInputBorder(
              borderSide: BorderSide(color: _kAccent, width: 2),
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Annulla',
                style: TextStyle(color: Colors.white54)),
          ),
          TextButton(
            onPressed: () {
              final name = controller.text.trim();
              if (name.isNotEmpty) {
                Navigator.of(context).pop();
                ref.read(equipmentProvider.notifier).selectCustom(name);
                widget.onComplete();
              }
            },
            child:
                const Text('Conferma', style: TextStyle(color: _kAccent)),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _kBackground,
      appBar: AppBar(
        backgroundColor: _kBackground,
        elevation: 0,
        title: const Text(
          'Seleziona dispositivo',
          style: TextStyle(color: _kOnSurface, fontWeight: FontWeight.w600),
        ),
        iconTheme: const IconThemeData(color: _kOnSurface),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator(color: _kAccent))
          : Column(
              children: [
                _SearchBar(controller: _searchController),
                Expanded(
                  child: ListView.builder(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 16, vertical: 8),
                    itemCount: _filtered.length + 1,
                    itemBuilder: (context, index) {
                      if (index < _filtered.length) {
                        return _EquipmentCard(
                          profile: _filtered[index],
                          onTap: () => _selectEquipment(_filtered[index]),
                        );
                      }
                      return _CustomEquipmentCard(onTap: _showCustomDialog);
                    },
                  ),
                ),
              ],
            ),
    );
  }
}

class _SearchBar extends StatelessWidget {
  final TextEditingController controller;

  const _SearchBar({required this.controller});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
      child: TextField(
        controller: controller,
        style: const TextStyle(color: _kOnSurface),
        cursorColor: _kAccent,
        decoration: InputDecoration(
          hintText: 'Cerca dispositivo…',
          hintStyle: const TextStyle(color: Colors.white38),
          prefixIcon: const Icon(Icons.search, color: Colors.white38),
          filled: true,
          fillColor: _kSurface,
          contentPadding:
              const EdgeInsets.symmetric(vertical: 0, horizontal: 16),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide.none,
          ),
        ),
      ),
    );
  }
}

class _EquipmentCard extends StatelessWidget {
  final EquipmentProfile profile;
  final VoidCallback onTap;

  const _EquipmentCard({required this.profile, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      color: _kSurface,
      margin: const EdgeInsets.symmetric(vertical: 6),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                profile.name,
                style: const TextStyle(
                  color: _kOnSurface,
                  fontWeight: FontWeight.w600,
                  fontSize: 15,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                profile.manufacturer,
                style: const TextStyle(
                  color: Colors.white54,
                  fontSize: 13,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _CustomEquipmentCard extends StatelessWidget {
  final VoidCallback onTap;

  const _CustomEquipmentCard({required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      color: _kSurface,
      margin: const EdgeInsets.symmetric(vertical: 6),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: _kAccent, width: 1),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: const Padding(
          padding: EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          child: Row(
            children: [
              Icon(Icons.add_circle_outline, color: _kAccent, size: 22),
              SizedBox(width: 12),
              Text(
                'Altro dispositivo',
                style: TextStyle(
                  color: _kAccent,
                  fontWeight: FontWeight.w600,
                  fontSize: 15,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
