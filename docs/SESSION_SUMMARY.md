# Session Summary

**Latest (2026-07-17):** Four low-cost vision routes were benchmarked; `google/gemma-3-4b-it:deepinfra` passed all menu/food/other cases and was selected at a live listed $0.05/M input and $0.10/M output tokens.
**Verification contract:** Release `fd2cc94` and migration 000028 require every discovered scraped photo to resolve and return a successful structured result before `ocr_status=verified`; partial counts, model, token use, and sanitized errors are persisted and shown in the restaurant UI.
**Production pilot:** One restaurant completed all 10/10 photos in 51 seconds using 6,110 input and 1,068 output tokens (about $0.00041), then created draft-only demo/campaign artifacts. Production is 1 verified, 8 failed, and 470 pending.
**Verified:** Go tests/vet/build, admin lint/TypeScript/Node 22 build, six new OCR tests, migration transaction test, deployment health, backup, and database constraint checks passed; 30/32 automation tests passed with only the two known retired-ingestion tests failing.
**Safety:** OCR remains persistently disabled and unscheduled, no OCR container is running, and no profile/demo/campaign approval, publication, or outreach send occurred.
