#!/bin/sh
# setup-extensions.sh
# Run ONCE when setting up a new database (requires superuser privileges).
# On managed databases (Supabase, RDS), install extensions via the dashboard instead.
#
# Usage:
#   SUPERUSER_DATABASE_URL=postgres://superuser:pass@host:5432/dbname ./scripts/setup-extensions.sh
#   or:
#   ./scripts/setup-extensions.sh postgres://superuser:pass@host:5432/dbname

set -e

DB_URL="${1:-$SUPERUSER_DATABASE_URL}"

if [ -z "$DB_URL" ]; then
  echo "Error: provide database URL as argument or set SUPERUSER_DATABASE_URL"
  exit 1
fi

echo "Installing PostgreSQL extensions..."

psql "$DB_URL" -c "CREATE EXTENSION IF NOT EXISTS pg_trgm;"
psql "$DB_URL" -c "CREATE EXTENSION IF NOT EXISTS unaccent;"  # useful for Indonesian text search

echo "Extensions installed successfully."
echo "You can now start the application — migrations will run automatically."
