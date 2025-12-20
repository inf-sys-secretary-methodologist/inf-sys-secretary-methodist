#!/bin/bash
# Entrypoint for backup container
# Supports: cron mode, one-shot mode, restore mode

set -euo pipefail

MODE="${BACKUP_MODE:-cron}"
BACKUP_SCHEDULE="${BACKUP_SCHEDULE:-0 2 * * *}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

case "${MODE}" in
    cron)
        log "Starting backup service in CRON mode"
        log "Schedule: ${BACKUP_SCHEDULE}"

        # Create cron job
        echo "${BACKUP_SCHEDULE} /scripts/backup-all.sh >> /var/log/backup.log 2>&1" > /etc/crontabs/root

        # Create log file
        touch /var/log/backup.log

        # Run initial backup if requested
        if [ "${BACKUP_ON_START:-false}" = "true" ]; then
            log "Running initial backup..."
            /scripts/backup-all.sh
        fi

        log "Starting cron daemon..."
        crond -f -l 2
        ;;

    backup)
        log "Starting one-shot backup"
        /scripts/backup-all.sh
        ;;

    backup-postgres)
        log "Starting PostgreSQL backup only"
        /scripts/backup-postgres.sh
        ;;

    backup-minio)
        log "Starting MinIO backup only"
        /scripts/backup-minio.sh
        ;;

    restore-postgres)
        log "Starting PostgreSQL restore"
        /scripts/restore-postgres.sh "$@"
        ;;

    restore-minio)
        log "Starting MinIO restore"
        /scripts/restore-minio.sh "$@"
        ;;

    shell)
        log "Starting shell"
        exec /bin/bash
        ;;

    *)
        log "Unknown mode: ${MODE}"
        log "Available modes: cron, backup, backup-postgres, backup-minio, restore-postgres, restore-minio, shell"
        exit 1
        ;;
esac
