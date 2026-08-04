# Andre Admin

Ops console for the Real Voice Agent: property listings CRUD + browser talk / outbound dial.

## Prerequisites

1. Voice agent running on `:8001` (`andre-voice-agent`)
2. Matching `CALL_API_SECRET` in agent `.env` and this app `.env.local`

## Run

```bash
cd apps/andre-admin
cp .env.example .env.local   # if needed
npm install
npm run dev
```

Open http://localhost:3003

### Default admin (local)

- Email: `admin@tuvi.local`
- Password: `andre-admin-123`

Session idle timeout: **10 minutes** (sliding). Unused sessions are logged out automatically.

## Pages

| Path | Purpose |
|------|---------|
| `/login` | Admin login |
| `/properties` | List / create / edit / delete inventory |
| `/voice` | Language, Start talking, place outbound call |
