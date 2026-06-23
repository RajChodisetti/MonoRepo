# OpenAPI / Swagger Documentation

Machine-readable API contract for the Phase 1 backend.

## Files

| File | Purpose |
|------|---------|
| `openapi.yaml` | OpenAPI 3.0 spec — all current endpoints |

## View in browser (Swagger UI)

### Option A — Swagger Editor (online)

1. Open [https://editor.swagger.io](https://editor.swagger.io)
2. **File → Import file** → select `docs/openapi/openapi.yaml`

### Option B — Local (Makefile)

```bash
make swagger
```

Open [http://localhost:8081](http://localhost:8081) (override port: `make swagger SWAGGER_PORT=9090`).

### Option C — Docker (manual)

```bash
docker run --rm -p 8081:8080 \
  -e SWAGGER_JSON=/openapi/openapi.yaml \
  -v "$(pwd)/docs/openapi:/openapi" \
  swaggerapi/swagger-ui
```

### Option D — VS Code / Cursor

Install extension **OpenAPI (Swagger) Editor** and open `openapi.yaml`.

## Validate

From repo root:

```bash
make openapi
```

## Keep in sync

When adding or changing routes in `backend/internal/http/router.go` or handlers:

1. Update `docs/openapi/openapi.yaml`
2. Run `make openapi`
3. Update `postman/` collection if needed

## Related

- Postman: `postman/README.md`
- Architecture + JSON examples: `docs/work-log/2026-06-23-phase1-api-architecture-diagram.md`
