# Knowledge Base — Appliance (Elettrodomestici)

Dominio: `appliance` | Problemi documentati: 20

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `lavatrice` | Lavatrice, asciugatrice, lavasciuga |
| `lavastoviglie` | Lavastoviglie da incasso e free-standing |
| `frigorifero` | Frigoriferi, freezer, cantinette |
| `forno` | Forni elettrici, microonde, forni a vapore |
| `cucina` | Piani cottura a gas e a induzione |
| `aspirapolvere` | Aspira polvere tradizionali e robot |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "lavatrice | lavastoviglie | frigorifero | forno | cucina | aspirapolvere",
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
        { "instruction": "Passo 1", "verification": "Come verificare l'esito", "time_seconds": 120 }
      ]
    }
  ]
}
```

## Come aggiungere un problema

1. Aprire `appliance_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Guasto che blocca il funzionamento o rischio sicurezza |
| `warning` | Anomalia che riduce le prestazioni |
| `info` | Manutenzione programmata |

## Note

- Tutti i testi sono in italiano
- Problemi a gas (cucine, riscaldamento): segnalare sempre `severity: error` e richiedere tecnico
- Le probabilità delle cause devono sommare circa 1.0
