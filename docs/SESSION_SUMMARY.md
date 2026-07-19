# Session Summary

Locally implemented migration `000036` and code changes so `demo_ready` is derived from `ocr_status='verified'` plus a nonblank restaurant email; manual demo creation no longer marks a lead demo-ready.
OCR/import finalization, restaurant email/status edits, and lead preparation now resynchronize only `lead`/`demo_ready` rows from that rule while preserving later lifecycle statuses.
Read-only production SQL showed schema `000035`, `verified=0`, `pending=940`, `failed=4`, and `demo_ready=22` but `eligible=0`; the OCR verified-only filter is empty because production has no verified profiles.
OCR worker logs still show the July 19 UTC daily request cap at `200/200`; deployment of this local fix and migration `000036` requires explicit approval.
