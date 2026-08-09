# Fine-tune dataset (scaffold)

Add labeled examples here for a future Gemini / Vertex fine-tune.

## Layout

```
dataset/
  images/          # meter_001.jpg, gas_001.jpg, ...
  labels/          # meter_001.json matching ElectricalAnalysis schema
```

Each label JSON should use the same fields as `app/schema.py` / `examples/meter_panel_sample.gold.json`.

## Suggested mix (10–50 images)
- Electrical meter boards (NMI visible / partially occluded)
- Fuse / switchboard only
- Gas pipes / meters with little or no text
- Mixed cabinets

## Build JSONL

```bash
python scripts/prepare_finetune_jsonl.py --out exports/finetune.jsonl
```

## Real fine-tune note
Gemini supervised fine-tuning typically runs via **Vertex AI** with a JSONL of multimodal examples and incurs training cost. This POC uses prompt + few-shot for day-one accuracy; grow `dataset/` before kicking off a tune job.
