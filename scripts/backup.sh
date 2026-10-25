#!/bin/sh
# Dumps the database every BACKUP_INTERVAL_SECONDS and prunes dumps older than
# BACKUP_KEEP_DAYS. Connection settings come from the PG* environment.
set -eu

: "${BACKUP_KEEP_DAYS:?}" "${BACKUP_INTERVAL_SECONDS:?}"

while :; do
  file="/backups/yupana-$(date -u +%Y%m%dT%H%M%SZ).dump"
  # Write to a temp name first, so a failed dump never looks like a backup.
  if pg_dump --format=custom --file="$file.partial"; then
    mv "$file.partial" "$file"
    echo "backup written: $file"
  else
    rm -f "$file.partial"
    echo "backup failed" >&2
  fi
  find /backups -name 'yupana-*.dump' -mtime +"$BACKUP_KEEP_DAYS" -print -delete
  sleep "$BACKUP_INTERVAL_SECONDS"
done
