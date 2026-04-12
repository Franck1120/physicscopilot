# Knowledge Base — HVAC

Dominio: `hvac` | Problemi documentati: 20

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `caldaia` | Caldaie murali e a condensazione |
| `climatizzatore` | Split, multi-split, inverter |
| `ventilazione` | VMC, estrattori, aerazione |
| `termostato` | Regolatori digitali e analogici |
| `radiatori` | Radiatori in alluminio e in ghisa |
| `pompa_di_calore` | Sistemi aria-acqua e geotermici |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "caldaia | climatizzatore | ...",
  "severity": "error | warning | info",
  "description": "Descrizione sintetica del problema.",
  "visual_symptoms": ["Sintomo visivo 1", "Sintomo visivo 2"],
  "probable_causes": [
    { "cause": "Causa principale", "probability": 0.35, "test": "Come verificare" }
  ],
  "solutions": [
    {
      "difficulty": "beginner | intermediate | expert",
      "steps": [
        { "instruction": "Passo 1", "verification": "Come verificare l'esito", "time_seconds": 60 }
      ]
    }
  ]
}
```

## Come aggiungere un problema

1. Aprire `hvac_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Guasto che impedisce il funzionamento |
| `warning` | Anomalia che riduce le prestazioni |
| `info` | Segnalazione di manutenzione ordinaria |

## Note

- Tutti i testi sono in italiano
- Le probabilità delle cause devono sommare circa 1.0
- `time_seconds` rappresenta il tempo stimato per eseguire il passo
