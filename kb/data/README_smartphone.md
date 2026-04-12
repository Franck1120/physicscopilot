# Knowledge Base — Smartphone

Dominio: `smartphone` | Problemi documentati: 20

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `display` | Schermo rotto, tocco non responsivo, burn-in |
| `batteria` | Scarica rapida, rigonfiamento, calibrazione |
| `connettivita` | Wi-Fi, Bluetooth, rete dati, NFC |
| `fotocamera` | Messa a fuoco, qualità foto, flash |
| `audio` | Microfono, altoparlante, auricolare |
| `hardware` | Pulsanti fisici, porte, vibrazione |
| `software` | App che crashano, aggiornamenti, storage |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "display | batteria | ...",
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

## Come aggiungere un problema

1. Aprire `smartphone_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Guasto che rende il dispositivo inutilizzabile |
| `warning` | Anomalia che limita le funzionalità |
| `info` | Segnalazione di usura normale |

## Note

- Tutti i testi sono in italiano
- Le probabilità delle cause devono sommare circa 1.0
- Includere sempre i sintomi visivi per supportare la diagnosi tramite fotocamera
