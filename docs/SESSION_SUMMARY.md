# Session Summary

Italian Villa experimental `template=4` is live for personalized restaurant demos, payload-driven through the existing lead/demo site contract with safe media, reservations, template switching, and the floating voice-agent button.
The admin restaurant Demo tab now lists “Italian Villa experimental” under Experimental templates, and demo engagement analytics accepts `template_id=4` via migration `000033`.
The VM is running app release `5ffddf2` at `/opt/tuvi/releases/monorepo-5ffddf2`; rollback points to `/opt/tuvi/releases/monorepo-3b0c246`, and the pre-migration backup is `/opt/tuvi/backups/pre-italian-template-3b0c246-20260719-185708.sql.gz`.
Local TypeScript, admin ESLint, Go tests, VM Docker builds, production HTTPS smokes, and zero-restart container checks passed; restaurants without safe verified media still render without image-heavy sections by design.
