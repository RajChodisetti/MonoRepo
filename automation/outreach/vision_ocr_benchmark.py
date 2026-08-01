#!/usr/bin/env python3
"""Read-only benchmark for low-cost vision routes used by restaurant OCR."""

from __future__ import annotations

import argparse
import json
import statistics
import time

from hf_llm import hf_api_key
from menu_image_ocr import MenuImageAnalyzer, MenuOCRConfig


CASES = (
    {
        "label": "restaurant_menu",
        "url": (
            "https://upload.wikimedia.org/wikipedia/commons/thumb/0/0d/"
            "Street_food_corner_menu.jpg/500px-Street_food_corner_menu.jpg"
        ),
        "expected_type": "menu_document",
        "min_items": 3,
    },
    {
        "label": "food_photo",
        "url": (
            "https://upload.wikimedia.org/wikipedia/commons/thumb/e/ec/"
            "Pizza_on_plate.jpg/330px-Pizza_on_plate.jpg"
        ),
        "expected_type": "food_photo",
        "min_items": 0,
    },
    {
        "label": "non_restaurant_photo",
        "url": (
            "https://cdn.britannica.com/61/93061-050-99147DCE/"
            "Statue-of-Liberty-Island-New-York-Bay.jpg"
        ),
        "expected_type": "other",
        "min_items": 0,
    },
)


def benchmark_model(model: str) -> dict:
    cfg = MenuOCRConfig(
        huggingface_api_key=hf_api_key(),
        hf_vision_model=model,
        openai_api_key="",
        gemini_api_key="",
        delay=0,
    )
    analyzer = MenuImageAnalyzer(cfg)
    cases: list[dict] = []
    latencies: list[float] = []
    input_tokens = 0
    output_tokens = 0

    for case in CASES:
        started = time.perf_counter()
        try:
            result = analyzer.analyze_image_url(case["url"])
        except Exception as exc:
            cases.append({
                "label": case["label"],
                "success": False,
                "correct_type": False,
                "items_extracted": 0,
                "error_type": type(exc).__name__,
            })
            continue

        latency = time.perf_counter() - started
        latencies.append(latency)
        usage = result.get("usage") or {}
        input_tokens += int(usage.get("input_tokens") or 0)
        output_tokens += int(usage.get("output_tokens") or 0)
        item_count = len(result.get("menu_items") or [])
        correct_type = result.get("image_type") == case["expected_type"]
        enough_items = item_count >= int(case["min_items"])
        cases.append({
            "label": case["label"],
            "success": True,
            "correct_type": correct_type,
            "enough_items": enough_items,
            "image_type": result.get("image_type"),
            "items_extracted": item_count,
            "latency_seconds": round(latency, 3),
        })

    passed = sum(
        1
        for case in cases
        if case.get("success") and case.get("correct_type") and case.get("enough_items")
    )
    return {
        "model": model,
        "cases_passed": passed,
        "cases_total": len(CASES),
        "successful_requests": sum(1 for case in cases if case.get("success")),
        "median_latency_seconds": (
            round(statistics.median(latencies), 3) if latencies else None
        ),
        "input_tokens": input_tokens,
        "output_tokens": output_tokens,
        "cases": cases,
    }


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Benchmark low-cost Hugging Face vision routes without database writes"
    )
    parser.add_argument(
        "--model",
        action="append",
        required=True,
        help="Full router model string, including an explicit provider suffix",
    )
    args = parser.parse_args()

    if not hf_api_key():
        raise SystemExit("HUGGING_FACE_API_KEY is required")

    results = [benchmark_model(model) for model in args.model]
    print(json.dumps({"benchmark": results}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
