#!/usr/bin/env bash
# Regenerate TypeScript SDK types from the OpenAPI spec.
# Requires: Node.js, npx (or openapi-typescript installed)
set -euo pipefail

SPEC="docs/swagger.json"
OUT="sdk/typescript/api.ts"

if [ ! -f "$SPEC" ]; then
  echo "❌ OpenAPI spec not found: $SPEC"
  echo "   Run 'swag init -g cmd/cosmoria/main.go -o docs/' first."
  exit 1
fi

echo "🔄 Generating TypeScript types from $SPEC..."
npx --yes openapi-typescript "$SPEC" -o "$OUT"
echo "✅ Generated $OUT"
