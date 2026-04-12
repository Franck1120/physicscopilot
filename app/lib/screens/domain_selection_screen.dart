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
/// Displays a 3-column grid of domain cards. A search field above the grid
/// filters domains by label (case-insensitive). Tapping a card saves the
/// chosen domain via [settingsProvider] and calls [onSelected].
class DomainSelectionScreen extends ConsumerStatefulWidget {
  const DomainSelectionScreen({super.key, this.onSelected});

  /// Called after the domain is persisted. Receives the selected domain id.
  final void Function(String domain)? onSelected;

  @override
  ConsumerState<DomainSelectionScreen> createState() =>
      _DomainSelectionScreenState();
}

class _DomainSelectionScreenState extends ConsumerState<DomainSelectionScreen> {
  final _searchController = TextEditingController();
  String _query = '';

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  List<_DomainItem> get _filtered {
    if (_query.isEmpty) return _domains;
    final lower = _query.toLowerCase();
    return _domains
        .where((d) => d.label.toLowerCase().contains(lower))
        .toList();
  }

  @override
  Widget build(BuildContext context) {
    final currentDomain = ref.watch(
      settingsProvider.select((s) => s.selectedDomain),
    );
    final filtered = _filtered;

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
              padding: const EdgeInsets.fromLTRB(20, 4, 20, 12),
              child: Text(
                'Seleziona il dominio per ottimizzare le analisi AI.',
                style: Theme.of(context)
                    .textTheme
                    .bodySmall
                    ?.copyWith(color: kTextMuted),
              ),
            ),

            // ── Search field ──────────────────────────────────────────────
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
              child: TextField(
                controller: _searchController,
                style: const TextStyle(color: Colors.white, fontSize: 14),
                decoration: InputDecoration(
                  hintText: 'Cerca dominio…',
                  hintStyle: const TextStyle(color: kTextMuted, fontSize: 14),
                  prefixIcon: const Icon(
                    Icons.search_rounded,
                    color: kTextMuted,
                    size: 20,
                  ),
                  suffixIcon: _query.isNotEmpty
                      ? IconButton(
                          icon: const Icon(
                            Icons.clear_rounded,
                            color: kTextMuted,
                            size: 18,
                          ),
                          onPressed: () {
                            _searchController.clear();
                            setState(() => _query = '');
                          },
                        )
                      : null,
                  filled: true,
                  fillColor: kBgCard,
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: 14,
                    vertical: 10,
                  ),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(10),
                    borderSide: const BorderSide(color: kBgCardBorder),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(10),
                    borderSide: const BorderSide(color: kBgCardBorder),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(10),
                    borderSide: const BorderSide(color: kAccent),
                  ),
                ),
                onChanged: (value) => setState(() => _query = value),
              ),
            ),

            // ── Domain grid ───────────────────────────────────────────────
            Expanded(
              child: filtered.isEmpty
                  ? const Center(
                      child: Text(
                        'Nessun dominio trovato',
                        style: TextStyle(color: kTextMuted, fontSize: 14),
                      ),
                    )
                  : GridView.builder(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      gridDelegate:
                          const SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: 3,
                        mainAxisSpacing: 12,
                        crossAxisSpacing: 12,
                        childAspectRatio: 0.9,
                      ),
                      itemCount: filtered.length,
                      itemBuilder: (context, index) {
                        final domain = filtered[index];
                        final isSelected = currentDomain == domain.id;
                        return _DomainCard(
                          item: domain,
                          selected: isSelected,
                          onTap: () async {
                            await ref
                                .read(settingsProvider.notifier)
                                .setDomain(domain.id);
                            widget.onSelected?.call(domain.id);
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
