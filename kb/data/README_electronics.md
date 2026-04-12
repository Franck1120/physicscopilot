# Knowledge Base — Electronics (Elettronica)

Dominio: `electronics` | Problemi documentati: 20

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `televisore` | TV LCD, OLED, QLED — display e audio |
| `audio` | Amplificatori, casse, soundbar |
| `console` | PlayStation, Xbox, Nintendo Switch |
| `proiettore` | Proiettori home cinema e da ufficio |
| `tablet` | iPad e Android tablet |
| `lettori` | Lettori Blu-ray, streaming stick |
| `alimentatori` | Adattatori, caricatori, UPS |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "televisore | audio | console | proiettore | tablet | lettori | alimentatori",
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

1. Aprire `electronics_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Il dispositivo non funziona o presenta rischio sicurezza |
| `warning` | Funzionalità ridotta o degrado qualità |
| `info` | Manutenzione e calibrazione |

## Note

- Tutti i testi sono in italiano
- Non aprire alimentatori sotto tensione: indicare sempre di staccare la spina
- Le probabilità delle cause devono sommare circa 1.0
