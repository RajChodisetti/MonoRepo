#!/usr/bin/env bash
cd /Users/praveenmaurya/Desktop/Tuvi/MonoRepo || exit 1
make db-up
echo ""
echo "API :8080 — keep this Terminal window open"
echo "================================================"
exec make api
