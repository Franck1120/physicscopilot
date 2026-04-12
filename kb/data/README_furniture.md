# Knowledge Base — Furniture (Mobili)

Dominio: `furniture` | Problemi documentati: 15

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `struttura` | Gambe rotte, articolazioni allentate, cedimenti |
| `rivestimento` | Tessuto strappato, pelle scollata, graffi |
| `meccanismi` | Cerniere, cassetti, sistemi reclinabili |
| `assemblaggio` | Viti allentate, pannelli spanciati, disallineamenti |
| `superfici` | Graffi, bruciature, macchie su legno/vetro |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "struttura | rivestimento | meccanismi | assemblaggio | superfici",
  "severity": "error | warning | info",
  "description": "Descrizione sintetica del problema.",
  "visual_symptoms": ["Sintomo visivo 1", "Sintomo visivo 2"],
  "probable_causes": [
    { "cause": "Causa principale", "probability": 0.50, "test": "Come verificare" }
  ],
  "solutions": [
    {
      "difficulty": "beginner | intermediate | expert",
      "steps": [
        { "instruction": "Passo 1", "verification": "Come verificare l'esito", "time_seconds": 300 }
      ]
    }
  ]
}
```

## Come aggiungere un problema

1. Aprire `furniture_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Problema di sicurezza o inutilizzabilità del mobile |
| `warning` | Difetto estetico o funzionale rilevante |
| `info` | Usura normale o manutenzione preventiva |

## Note

- Tutti i testi sono in italiano
- Le probabilità delle cause devono sommare circa 1.0
- Per problemi strutturali con rischio sicurezza, usare sempre `severity: error`
