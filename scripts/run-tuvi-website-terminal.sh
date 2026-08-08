#!/usr/bin/env bash
cd /Users/praveenmaurya/Desktop/Tuvi/MonoRepo/web || exit 1
echo ""
echo "Tuvi website :3001 — keep this Terminal window open"
echo "Open: http://localhost:3001"
echo "================================================"
exec npm run dev -- -p 3001
