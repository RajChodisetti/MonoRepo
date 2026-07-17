# Session Summary

**Latest (2026-07-17):** Release `199241c` is on remote `master` and production; restaurant detail now shows fresh attributed Google photo URLs, stored OCR image URLs, explicit OCR checked/attempt/timestamp/error state, and the existing UUID-mapped website generator templates.
**Demo behavior:** Manual draft creation now snapshots the real restaurant payload; inspect is read-only, publish requires OCR verified plus human profile approval, and unpublish revokes public access while keeping the draft.
**Verified:** 157 Go tests plus vet/build, ESLint, TypeScript, Node 22 production build, OpenAPI/Compose validation, protected-route and production service checks passed; a valid pre-OCR backup and rollback images were created.
**OCR result:** A three-row controlled batch resolved ten Google images per row, then failed because Hugging Face no longer serves configured `Qwen/Qwen2-VL-7B-Instruct`; 3 rows show failed/checked, 234 remain pending, and no images were persisted.
**Approval needed:** OCR remains disabled and unscheduled until a current production vision route (for example `Qwen/Qwen3-VL-30B-A3B-Instruct`) is explicitly approved; outreach sending remains disabled.
