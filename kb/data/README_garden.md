# Knowledge Base — Garden (Giardino)

Dominio: `garden` | Problemi documentati: 15

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `tagliaerba` | Tagliaerba a benzina, elettrici e robot |
| `irrigazione` | Impianti a goccia, irrigatori, timer |
| `utensili` | Tosasiepi, soffiatori, motoseghe |
| `motopompe` | Pompe di superficie e sommerse |
| `generatori` | Generatori portatili da giardino |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "tagliaerba | irrigazione | utensili | motopompe | generatori",
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
        { "instruction": "Passo 1", "verification": "Come verificare l'esito", "time_seconds": 180 }
      ]
    }
  ]
}
```

## Come aggiungere un problema

1. Aprire `garden_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Guasto che blocca il funzionamento |
| `warning` | Anomalia che riduce le prestazioni |
| `info` | Manutenzione di stagione |

## Note

- Tutti i testi sono in italiano
- Includere sempre avvertenze di sicurezza nei passi che coinvolgono parti in movimento
- Le probabilità delle cause devono sommare circa 1.0
