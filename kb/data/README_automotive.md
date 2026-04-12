# Knowledge Base — Automotive (Auto)

Dominio: `automotive` | Problemi documentati: 20

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `motore` | Avvio difficile, perdite olio, surriscaldamento |
| `freni` | Freni a disco, a tamburo, freno a mano |
| `trasmissione` | Frizione, cambio, differenziale |
| `elettrica` | Batteria, alternatore, fusibili |
| `sospensioni` | Ammortizzatori, molle, bracci |
| `pneumatici` | Forature, usura irregolare, pressione |
| `carrozzeria` | Ammaccature, ruggine, parabrezza |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "motore | freni | trasmissione | elettrica | sospensioni | pneumatici | carrozzeria",
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
        { "instruction": "Passo 1", "verification": "Come verificare l'esito", "time_seconds": 300 }
      ]
    }
  ]
}
```

## Come aggiungere un problema

1. Aprire `automotive_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Il veicolo non è guidabile in sicurezza — non muoversi |
| `warning` | Anomalia che richiede verifica prima del prossimo viaggio |
| `info` | Tagliando, manutenzione programmata |

## Note

- Tutti i testi sono in italiano
- Problemi a freni e pneumatici: sempre `severity: error`
- Le probabilità delle cause devono sommare circa 1.0
