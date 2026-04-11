import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/printer_provider.dart';

const _kAccent = Color(0xFF1B4F72);
const _kBackground = Color(0xFF0D1B2A);
const _kSurface = Color(0xFF1A2B3C);
const _kOnSurface = Color(0xFFE0E8F0);

class PrinterSelectionScreen extends ConsumerStatefulWidget {
  final VoidCallback onComplete;

  const PrinterSelectionScreen({super.key, required this.onComplete});

  @override
  ConsumerState<PrinterSelectionScreen> createState() =>
      _PrinterSelectionScreenState();
}

class _PrinterSelectionScreenState
    extends ConsumerState<PrinterSelectionScreen> {
  List<PrinterProfile> _allProfiles = [];
  List<PrinterProfile> _filtered = [];
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
        .map((e) => PrinterProfile.fromJson(e as Map<String, dynamic>))
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

  void _selectPrinter(PrinterProfile printer) {
    ref.read(printerProvider.notifier).select(printer);
    widget.onComplete();
  }

  Future<void> _showCustomDialog() async {
    final controller = TextEditingController();
    await showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: _kSurface,
        title: const Text(
          'Nome stampante',
          style: TextStyle(color: _kOnSurface),
        ),
        content: TextField(
          controller: controller,
          autofocus: true,
          style: const TextStyle(color: _kOnSurface),
          cursorColor: _kAccent,
          decoration: const InputDecoration(
            hintText: 'Es. Ender 3 custom modded',
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
                ref
                    .read(printerProvider.notifier)
                    .selectCustom(name);
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
          'Seleziona stampante',
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
                        return _PrinterCard(
                          profile: _filtered[index],
                          onTap: () => _selectPrinter(_filtered[index]),
                        );
                      }
                      return _CustomPrinterCard(onTap: _showCustomDialog);
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
          hintText: 'Cerca stampante...',
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

class _PrinterCard extends StatelessWidget {
  final PrinterProfile profile;
  final VoidCallback onTap;

  const _PrinterCard({required this.profile, required this.onTap});

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
          child: Row(
            children: [
              Expanded(
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
              _ExtruderBadge(extruderType: profile.extruderType),
            ],
          ),
        ),
      ),
    );
  }
}

class _ExtruderBadge extends StatelessWidget {
  final String extruderType;

  const _ExtruderBadge({required this.extruderType});

  @override
  Widget build(BuildContext context) {
    final isDirect = extruderType == 'direct_drive';
    final label = isDirect ? 'Direct' : 'Bowden';
    final color = isDirect ? _kAccent : const Color(0xFF2E6B4E);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: color.withAlpha(51),
        border: Border.all(color: color.withAlpha(128)),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: isDirect ? const Color(0xFF5DADE2) : const Color(0xFF58D68D),
          fontSize: 12,
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }
}

class _CustomPrinterCard extends StatelessWidget {
  final VoidCallback onTap;

  const _CustomPrinterCard({required this.onTap});

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
                'Altra stampante',
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
