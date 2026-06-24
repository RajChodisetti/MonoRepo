# Postman — Restaurant Platform API

Import these files into Postman to test the full Phase 1 backend flow locally.

## Files

| File | Purpose |
|------|---------|
| `Restaurant-Platform.postman_collection.json` | All API requests + PM demo flow |
| `Restaurant-Platform-Local.postman_environment.json` | Local variables (URLs, credentials, auto-saved IDs) |

## Import

1. Open Postman → **Import** → select both JSON files
2. Top-right dropdown → **Restaurant Platform - Local**

## Terminal setup (run once)

```bash
cd MonoRepo
cp .env.example .env
make db-up
make migrate-up
make seed-admin
make api
```

API: `http://localhost:8080`

## PM demo flow (recommended)

In Postman, open folder **`01 — PM Demo Flow (in order)`** and run each request top to bottom (or use **Run collection** on that folder only).

| Step | What it proves |
|------|----------------|
| 1 Login Admin | Internal sales admin can authenticate |
| 2 Admin Me | Role = `internal_admin` |
| 3 Create Restaurant | Lead capture (name + email) |
| 4–5 List / Get | Admin sees restaurant data |
| 6 Mark Contacted | Outreach tracking → status `emailed` |
| 7 Set demo_ready | Lifecycle status management |
| 8 Signup Owner | Restaurant owner account |
| 9 Add Member | Tenant scoping — owner linked to restaurant |
| 10 Create Demo Site | Personalized demo link + one-time token |
| 11 Public Demo | Owner-facing demo (no login, safe payload) |
| 12 Owner List | Owner sees only their restaurant |
| 13 Shown Interest | Click tracking → status `interested` |
| 14 List Members | Membership visible to admin |
| 15 Archive | Soft delete / cleanup |

## Alternative: one-command fixture

```bash
make seed-demo-fixture
```

Then use folder **`08 — Quick Fixture`** with default `demo_slug` / `demo_token` in environment.

## Collection folders

| Folder | Contents |
|--------|----------|
| `01 — PM Demo Flow` | End-to-end sales MVP story |
| `02 — Auth` | Signup/login, Auth Me, negative tests |
| `03 — Admin` | Admin profile |
| `04 — Restaurants` | CRUD + query filters |
| `05 — Demo Sites` | Admin creates demo |
| `06 — Public Demo` | Public slug+token access |
| `07 — Health` | Developer-only health checks |
| `08 — Quick Fixture` | After `make seed-demo-fixture` |

## Environment variables (auto-filled by tests)

| Variable | Set by |
|----------|--------|
| `access_token` | Login / Signup scripts |
| `admin_access_token` | Admin login |
| `owner_access_token` | Owner signup/login |
| `restaurant_id` | Create Restaurant |
| `owner_user_id` | Owner signup/login |
| `demo_slug`, `demo_token` | Create Demo Site |

## Roles reference

| Role | Value | How |
|------|-------|-----|
| Internal admin | `internal_admin` | `make seed-admin` |
| Restaurant owner | `restaurant_owner` | Signup |
| Developer | `developer` | Signup (local env only) |

## Restaurant status values

`lead` · `demo_ready` · `emailed` · `interested` · `client_onboarding` · `active_client` · `lost` · `archived`

## List query params

`GET /api/v1/restaurants`

- `?restaurant=` — name substring
- `?status=` — lifecycle filter
- `?is_contacted=true|false`
- `?shown_interest=true|false`
- `?include_archived=true`

## OpenAPI / Swagger

Machine-readable spec: [`../docs/openapi/openapi.yaml`](../docs/openapi/openapi.yaml)

Validate: `make openapi` from repo root.

View UI: `make swagger` → http://localhost:8081

Guide: [`../docs/openapi/README.md`](../docs/openapi/README.md)
