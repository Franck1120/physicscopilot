# Knowledge Base — Computer

Dominio: `computer` | Problemi documentati: 20

## Categorie coperte

| Categoria | Descrizione |
|-----------|-------------|
| `hardware` | CPU, RAM, scheda madre, alimentatore |
| `storage` | HDD, SSD, NVMe — errori e lentezza |
| `display` | Monitor, scheda video, cavi |
| `periferiche` | Tastiera, mouse, stampante, USB |
| `rete` | Scheda di rete, driver Wi-Fi, Ethernet |
| `raffreddamento` | Ventole, pasta termica, dissipatori |
| `software` | BSOD, driver, boot fallito |

## Schema JSON

Ogni problema segue la struttura standard `KBEntry`:

```json
{
  "id": "identificatore_unico",
  "name": "Nome leggibile del problema",
  "category": "hardware | storage | display | periferiche | rete | raffreddamento | software",
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

1. Aprire `computer_problems.json`
2. Aggiungere un oggetto nell'array `problems` seguendo lo schema sopra
3. Incrementare `total` di 1
4. Aggiornare `last_updated` con la data corrente (`YYYY-MM-DD`)
5. Assicurarsi che `id` sia univoco nel file

## Severity

| Valore | Significato |
|--------|-------------|
| `error` | Il computer non si avvia o si spegne inaspettatamente |
| `warning` | Prestazioni degradate o instabilità |
| `info` | Manutenzione o ottimizzazione preventiva |

## Note

- Tutti i testi sono in italiano
- Per problemi di storage: includere sempre il backup come primo passo
- Le probabilità delle cause devono sommare circa 1.0
