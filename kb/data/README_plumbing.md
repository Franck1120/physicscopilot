# Knowledge Base — Plumbing (Idraulica)

Dominio: `plumbing` | Problemi documentati: 20

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `perdite` | Perdite da rubinetti, giunti, sifoni |
| `scarichi` | Ostruzioni lavello, doccia, WC |
| `wc` | Cassetta, galleggiante, sifone WC |
| `boiler` | Scaldacqua elettrici e a gas |
| `pressione` | Bassa o alta pressione in impianto |
| `tubazioni` | Tubi congelati, ruggine, calcare |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "perdite | scarichi | wc | boiler | pressione | tubazioni",
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
        { "instruction": "Passo 1", "verification": "Come verificare l'esito", "time_seconds": 300 }
      ]
    }
  ]
}
```

## Come aggiungere un problema

1. Aprire `plumbing_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Perdita attiva o rischio allagamento — intervento immediato |
| `warning` | Anomalia che può peggiorare nel tempo |
| `info` | Manutenzione preventiva (anticalcare, guarnizioni) |

## Note

- Tutti i testi sono in italiano
- Per perdite attive: il primo passo è sempre chiudere il rubinetto di arresto
- Le probabilità delle cause devono sommare circa 1.0
