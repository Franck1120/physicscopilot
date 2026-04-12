// Copyright (c) 2026 PhysicsCopilot. All rights reserved.
// SPDX-License-Identifier: MIT

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../providers/settings_provider.dart';
import '../../main.dart'
    show kAccent, kBgCard, kBgCardBorder, kBgPrimary, kTextMuted;

// ── Domain data ───────────────────────────────────────────────────────────────

class _DomainItem {
  final String id;
  final IconData icon;
  final String label;

  const _DomainItem(this.id, this.icon, this.label);
}

const _domains = [
  _DomainItem('printer', Icons.print, 'Stampanti'),
  _DomainItem('appliances', Icons.kitchen, 'Elettrodomestici'),
  _DomainItem('automotive', Icons.directions_car, 'Auto'),
  _DomainItem('hvac', Icons.ac_unit, 'HVAC'),
  _DomainItem('electronics', Icons.electrical_services, 'Elettronica'),
  _DomainItem('computer', Icons.computer, 'Computer'),
  _DomainItem('plumbing', Icons.plumbing, 'Idraulica'),
  _DomainItem('bicycle', Icons.directions_bike, 'Biciclette'),
  _DomainItem('smartphone', Icons.smartphone, 'Smartphone'),
  _DomainItem('furniture', Icons.chair, 'Mobili'),
  _DomainItem('garden', Icons.yard, 'Giardino'),
  _DomainItem('photography', Icons.camera_alt, 'Fotografia'),
];

// ── Screen ────────────────────────────────────────────────────────────────────

/// Route: `/domain-selection`
///
/// Displays a 3-column grid of domain cards. Tapping a card saves the chosen
/// domain via [settingsProvider] and calls [onSelected] with the domain id.
class DomainSelectionScreen extends ConsumerWidget {
  const DomainSelectionScreen({super.key, this.onSelected});

  /// Called after the domain is persisted. Receives the selected domain id.
  final void Function(String domain)? onSelected;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final currentDomain = ref.watch(
      settingsProvider.select((s) => s.selectedDomain),
    );

    return Scaffold(
      backgroundColor: kBgPrimary,
      appBar: AppBar(
        title: const Text('Seleziona dominio'),
        backgroundColor: const Color(0xFF111111),
      ),
      body: SafeArea(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 20, 20, 4),
              child: Text(
                'In quale settore lavori?',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      color: Colors.white,
                      fontWeight: FontWeight.bold,
                    ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 4, 20, 16),
              child: Text(
                'Seleziona il dominio per ottimizzare le analisi AI.',
                style: Theme.of(context)
                    .textTheme
                    .bodySmall
                    ?.copyWith(color: kTextMuted),
              ),
            ),
            Expanded(
              child: GridView.builder(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 3,
                  mainAxisSpacing: 12,
                  crossAxisSpacing: 12,
                  childAspectRatio: 0.9,
                ),
                itemCount: _domains.length,
                itemBuilder: (context, index) {
                  final domain = _domains[index];
                  final isSelected = currentDomain == domain.id;
                  return _DomainCard(
                    item: domain,
                    selected: isSelected,
                    onTap: () async {
                      await ref
                          .read(settingsProvider.notifier)
                          .setDomain(domain.id);
                      onSelected?.call(domain.id);
                    },
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ── Domain card ───────────────────────────────────────────────────────────────

class _DomainCard extends StatelessWidget {
  const _DomainCard({
    required this.item,
    required this.selected,
    required this.onTap,
  });

  final _DomainItem item;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeInOut,
        decoration: BoxDecoration(
          color: selected ? kAccent.withAlpha(30) : kBgCard,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: selected ? kAccent : kBgCardBorder,
            width: selected ? 1.5 : 1,
          ),
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              item.icon,
              size: 32,
              color: selected ? kAccent : Colors.white70,
            ),
            const SizedBox(height: 8),
            Text(
              item.label,
              style: TextStyle(
                color: selected ? kAccent : Colors.white,
                fontSize: 12,
                fontWeight:
                    selected ? FontWeight.w600 : FontWeight.w400,
                height: 1.2,
              ),
              textAlign: TextAlign.center,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
      ),
    );
  }
}
