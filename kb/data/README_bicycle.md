# Knowledge Base — Bicycle (Biciclette)

Dominio: `bicycle` | Problemi documentati: 20

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `freni` | Freni a disco, freni a V, caliper |
| `trasmissione` | Catena, deragliatore, cassetta |
| `ruote` | Foratura, raggi rotti, cerchio storto |
| `sterzo` | Manubrio, attacco, cuscinetti di sterzo |
| `sella` | Reggisella, sella, ammortizzatori |
| `elettronica` | Bici elettrica, batteria, motore |
| `telaio` | Crepe, ammaccature, verniciatura |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "freni | trasmissione | ruote | sterzo | sella | elettronica | telaio",
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
        { "instruction": "Passo 1", "verification": "Come verificare l'esito", "time_seconds": 300 }
      ]
    }
  ]
}
```

## Come aggiungere un problema

1. Aprire `bicycle_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Problema di sicurezza — non pedalare finché non risolto |
| `warning` | Riduzione prestazioni o comfort |
| `info` | Manutenzione ordinaria e lubrificazione |

## Note

- Tutti i testi sono in italiano
- Problemi a freni e sterzo: usare sempre `severity: error`
- Le probabilità delle cause devono sommare circa 1.0
