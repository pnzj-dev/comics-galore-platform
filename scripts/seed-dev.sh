#!/bin/bash
# scripts/seed-dev.sh — one command to seed all dev data (idempotent)
# Usage: ./scripts/seed-dev.sh [endpoint] [token]
#   Default: http://localhost:4000 dev-secret

set -e

ENDPOINT="${1:-http://localhost:4000}"
TOKEN="${2:-dev-secret}"

echo "Seeding dev data at $ENDPOINT ..."
echo ""

echo "--- Users ---"
curl -s -X POST "$ENDPOINT/dev/seed-users" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}" | python3 -m json.tool 2>/dev/null || echo "(json.tool not available, raw output above)"

echo ""
echo "--- Comics ---"
curl -s -X POST "$ENDPOINT/dev/seed-comics" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}" | python3 -m json.tool 2>/dev/null || echo "(json.tool not available, raw output above)"

echo ""
echo "--- Series ---"
curl -s -X POST "$ENDPOINT/dev/seed-series" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}" | python3 -m json.tool 2>/dev/null || echo "(json.tool not available, raw output above)"

echo ""
echo "--- Engagement ---"
curl -s -X POST "$ENDPOINT/dev/seed-engagement" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}" | python3 -m json.tool 2>/dev/null || echo "(json.tool not available, raw output above)"

echo ""
echo "--- Billing ---"
curl -s -X POST "$ENDPOINT/dev/seed-billing" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}" | python3 -m json.tool 2>/dev/null || echo "(json.tool not available, raw output above)"

echo ""
echo "--- Quota (exhaust free-tier downloads for boost testing) ---"
curl -s -X POST "$ENDPOINT/dev/seed-quota" \
  -H "Content-Type: application/json" \
  -d "{\"token\":\"$TOKEN\"}" | python3 -m json.tool 2>/dev/null || echo "(json.tool not available, raw output above)"

echo ""
echo "Seed complete."
echo ""
echo "Demo users (password: devpassword):"
echo "  admin@comics-galore.dev           (admin / platinum)"
echo "  author-free@pnzj.dev              (uploader / free)"
echo "  author-gold@pnzj.dev              (uploader / gold)"
echo "  member-free@pnzj.dev              (user / free)"
echo "  member-bronze@pnzj.dev            (user / bronze)"
echo "  member-silver@pnzj.dev            (user / silver)"
echo "  member-gold@pnzj.dev              (user / gold)"
echo "  member-platinum@pnzj.dev          (user / platinum)"
echo "  member-exhausted@pnzj.dev         (user / free — download quota exhausted)"
