#!/bin/bash
# Launches the Tuvi website dev server (kills any stale one first) and opens the browser.
export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"
[ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh"

cd "$(dirname "$0")/web" || exit 1

# Kill any stale/hung dev server holding port 3001
pids="$(lsof -nP -tiTCP:3001 -sTCP:LISTEN 2>/dev/null)"
if [ -n "$pids" ]; then
  echo "Stopping old server (pid $pids)..."
  kill -9 $pids 2>/dev/null
  sleep 1
fi

# Open the browser once the server is up
( sleep 8 && open "http://localhost:3001" ) &

echo "Starting Tuvi website on http://localhost:3001 ..."
npm run dev -- -p 3001
