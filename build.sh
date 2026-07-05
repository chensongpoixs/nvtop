#!/bin/bash
set -e

BASEDIR="$(cd "$(dirname "$0")" && pwd)"
GOPATH="${GOPATH:-$(go env GOPATH 2>/dev/null || echo $HOME/go)}"
GO="${GOPATH}/bin/go"
if [ ! -x "$GO" ]; then
	GO="$(which go 2>/dev/null || echo go)"
fi

echo "==> Building Vue frontend..."
cd "$BASEDIR/frontend"
npm install --silent
npm run build

echo "==> Copying frontend dist to backend..."
rm -rf "$BASEDIR/backend/frontend/dist"
cp -r "$BASEDIR/frontend/dist" "$BASEDIR/backend/frontend/dist"

echo "==> Building Go server..."
CGO_ENABLED=1 "$GO" build -C "$BASEDIR/backend" -o "$BASEDIR/nvtop-server" .

echo "==> Build complete: $BASEDIR/nvtop-server"
ls -lh "$BASEDIR/nvtop-server"
