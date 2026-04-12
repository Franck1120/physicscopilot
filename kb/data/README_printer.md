# Knowledge Base — Printer (Stampanti)

Dominio: `printer` | Problemi documentati: 20+

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `estrusione` | Ugello ostruito, under/over-extrusion |
| `adesione` | Warping, prima passata, bed leveling |
| `qualita_stampa` | Stringing, layer shift, gap tra layers |
| `meccanica` | Motori passo-passo, cinghie, guide |
| `temperatura` | Hotend, bed riscaldato, thermal runaway |
| `connettivita` | USB, Wi-Fi, carta bloccata (2D) |
| `inchiostro` | Cartucce, testa stampa, levatura (2D) |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "estrusione | adesione | qualita_stampa | meccanica | temperatura | connettivita | inchiostro",
  "severity": "error | warning | info",
  "description": "Descrizione sintetica del problema.",
  "visual_symptoms": ["Sintomo visivo 1", "Sintomo visivo 2"],
  "probable_causes": [
    { "cause": "Causa principale", "probability": 0.40, "test": "Come verificare" }
  ],
  "solutions": [
    {
      "difficulty": "beginner | intermediate | expert",
      "steps": [
        { "instruction": "Passo 1", "verification": "Come verificare l'esito", "time_seconds": 120 }
      ]
    }
  ]
}
```

## File disponibili

| File | Contenuto |
|------|-----------|
| `problems.json` | Problemi generici di stampa 3D |
| `printer_profiles.json` | Profili specifici per modello di stampante |

## Come aggiungere un problema

1. Aprire `problems.json` o `printer_profiles.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | La stampa fallisce o la stampante è inutilizzabile |
| `warning` | Qualità di stampa ridotta |
| `info` | Calibrazione e manutenzione preventiva |

## Note

- Tutti i testi sono in italiano
- Le probabilità delle cause devono sommare circa 1.0
- Includere sempre la temperatura consigliata per filamento/carta nei passi di soluzione
