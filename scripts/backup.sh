#!/bin/bash
# Quick backup script - run from project root
# Usage: ./scripts/backup.sh [postgres|minio|all]

set -euo pipefail

cd "$(dirname "$0")/.."

TYPE="${1:-all}"

case "${TYPE}" in
    postgres)
        echo "Running PostgreSQL backup..."
        docker compose run --rm backup /scripts/backup-postgres.sh
        ;;
    minio)
        echo "Running MinIO backup..."
        docker compose run --rm backup /scripts/backup-minio.sh
        ;;
    all)
        echo "Running full backup..."
        docker compose run --rm backup /scripts/backup-all.sh
        ;;
    *)
        echo "Usage: $0 [postgres|minio|all]"
        exit 1
        ;;
esac
