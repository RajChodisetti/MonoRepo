# Session Summary

Production runtime is `e4d6801` at `/opt/tuvi/releases/monorepo-e4d6801`; rollback points to `/opt/tuvi/releases/monorepo-88b58eb`, and schema remains at `000035` with no new migration.
The admin portal now has `/admin/developer`, a protected internal-admin Developer tab with a read-only SQL runner, 200-row capped results, schema table/column browser, and menu-item popularity shortcuts.
Restaurants and menus are stored in PostgreSQL: live counts after deployment showed 944 restaurants, 944 menus, 47 menu item rows, and 47 distinct normalized menu item names.
Backend tests/build/vet, admin lint/TypeScript, Node 22 admin production build, OpenAPI validation, VM image builds, deployment, and HTTPS protection smokes passed; Docker was unavailable locally, so the production-equivalent Docker build was verified on the VM.
