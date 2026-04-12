# Knowledge Base — Photography (Fotografia)

Dominio: `photography` | Problemi documentati: 15

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `obiettivo` | Messa a fuoco, aberrazioni, diaframma bloccato |
| `sensore` | Macchie, dead pixel, sensore sporco |
| `corpo_macchina` | Otturatore, mirino, baionetta |
| `flash` | Flash integrato e speedlight |
| `memoria` | Schede SD, errori lettura/scrittura |
| `batteria` | Autonomia, carica, batterie gonfie |
| `stabilizzazione` | IS/VR difettoso, shake blur |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "obiettivo | sensore | corpo_macchina | flash | memoria | batteria | stabilizzazione",
  "severity": "error | warning | info",
  "description": "Descrizione sintetica del problema.",
  "visual_symptoms": ["Sintomo visivo 1", "Sintomo visivo 2"],
  "probable_causes": [
    { "cause": "Causa principale", "probability": 0.45, "test": "Come verificare" }
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

## Come aggiungere un problema

1. Aprire `photography_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Il problema impedisce di scattare foto |
| `warning` | Il problema degradisce la qualità delle immagini |
| `info` | Manutenzione preventiva raccomandata |

## Note

- Tutti i testi sono in italiano
- I sintomi visivi devono descrivere cosa appare nelle foto o sul display
- Per problemi al sensore: raccomandare sempre la pulizia professionale come opzione `expert`
