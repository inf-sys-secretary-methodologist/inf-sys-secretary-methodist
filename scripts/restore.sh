#!/bin/bash
# Quick restore script - run from project root
# Usage: ./scripts/restore.sh [postgres|minio] [backup_file]

set -euo pipefail

cd "$(dirname "$0")/.."

TYPE="${1:-}"
BACKUP_FILE="${2:-}"

if [ -z "${TYPE}" ]; then
    echo "Usage: $0 <postgres|minio> [backup_file]"
    echo ""
    echo "Examples:"
    echo "  $0 postgres                    # Restore latest PostgreSQL backup"
    echo "  $0 minio                       # Restore latest MinIO backup"
    echo "  $0 postgres /path/to/backup    # Restore specific backup"
    exit 1
fi

echo "WARNING: This will overwrite current data!"
read -p "Are you sure? (yes/no): " CONFIRM
if [ "${CONFIRM}" != "yes" ]; then
    echo "Aborted"
    exit 0
fi

case "${TYPE}" in
    postgres)
        echo "Restoring PostgreSQL..."
        if [ -n "${BACKUP_FILE}" ]; then
            docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-postgres.sh "${BACKUP_FILE}"
        else
            docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-postgres.sh
        fi
        ;;
    minio)
        echo "Restoring MinIO..."
        if [ -n "${BACKUP_FILE}" ]; then
            docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-minio.sh "${BACKUP_FILE}"
        else
            docker compose run --rm -e RESTORE_CONFIRM=true backup /scripts/restore-minio.sh
        fi
        ;;
    *)
        echo "Unknown type: ${TYPE}"
        echo "Usage: $0 <postgres|minio> [backup_file]"
        exit 1
        ;;
esac

echo "Restore completed. Consider restarting backend:"
echo "  docker compose restart backend"
