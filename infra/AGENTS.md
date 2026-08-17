# Infrastructure Scope Instructions

## Scope

Applies to `infra/**`: Dockerfiles, local/VM Compose topology, proxy examples,
build arguments, runtime environment forwarding, and service health wiring.

## Current topology

- Root Go API/worker/migrate binaries build from `infra/docker/Dockerfile.backend`.
- Admin, restaurant template, corporate site, and static catalog are independent
  Node build units. Runtime-aligned Node is version 22.
- The durable scrape/import worker uses a Python 3.12 image and shares
  PostgreSQL contracts with Go. Production image OCR is retired.
- `docker-compose.yml` is local/profile-based; `docker-compose.vm.yml` is the VM
  deployment definition. Inspect both when shared service/config behavior changes.
- Caddy/proxy examples must preserve public versus private route boundaries.

## Dependency impact

- Build arg or env changes require the service's typed config/example env and
  all Compose forwarding sites to stay aligned.
- Port/base-path changes require `docs/SERVICES.md`, health checks, proxy config,
  frontend config, and scripts to be inspected.
- Schema-dependent services must use the migration job before API/workers start.
- Changing a Docker language/runtime version requires the matching local tool
  guidance and production build checks.
- Do not add a new service when the modular monolith or an existing worker can
  own the behavior.

## Safety

- Never bake credentials into images, Compose, examples, or build args.
- Do not expose PostgreSQL, Redis, private admin routes, or internal worker ports
  publicly.
- Production deploys, migrations, infrastructure, permissions, billing, DNS,
  certificates, and secret changes require explicit approval.
- Compose rendering and image builds are local checks; they do not authorize
  deployment or provider calls.
- Preserve rollback paths and the previous release for deployment proposals.

## Verification

```bash
rtk docker compose -f infra/docker/docker-compose.yml --profile stack config --quiet
rtk proxy env DATABASE_URL=postgres://example:example@localhost/example POSTGRES_PASSWORD=example TUVI_API_TOKEN=example CALL_API_SECRET=example docker compose -f infra/docker/docker-compose.vm.yml config --quiet
rtk git diff --check
```

Build only affected images unless a topology change requires the full stack.
Report missing protected env inputs without printing them.
