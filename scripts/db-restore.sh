#!/bin/bash
# Database restore script
# Usage: ./scripts/db-restore.sh <backup_file>

set -euo pipefail

BACKUP_FILE="${1:-}"
CONTAINER_NAME="${DB_CONTAINER:-inf-sys-postgres}"
DB_NAME="${POSTGRES_DB:-inf_sys_db}"
DB_USER="${POSTGRES_USER:-postgres}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Validate input
if [ -z "$BACKUP_FILE" ]; then
    log_error "Usage: $0 <backup_file>"
    log_info "Available backups:"
    ls -la ./backups/*.sql.gz 2>/dev/null || echo "No backups found in ./backups/"
    exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
    log_error "Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    log_error "PostgreSQL container '$CONTAINER_NAME' is not running"
    exit 1
fi

# Confirm restore
log_warn "This will restore database '$DB_NAME' from: $BACKUP_FILE"
log_warn "All current data will be OVERWRITTEN!"
read -p "Are you sure you want to continue? (yes/no): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    log_info "Restore cancelled"
    exit 0
fi

# Create a safety backup before restore
log_info "Creating safety backup before restore..."
SAFETY_BACKUP="./backups/pre-restore-$(date +%Y%m%d-%H%M%S).sql.gz"
mkdir -p ./backups
docker exec "$CONTAINER_NAME" pg_dump -U "$DB_USER" -d "$DB_NAME" | gzip > "$SAFETY_BACKUP"
log_info "Safety backup created: $SAFETY_BACKUP"

# Restore database
log_info "Restoring database from: $BACKUP_FILE"

# Disconnect all active connections
docker exec "$CONTAINER_NAME" psql -U "$DB_USER" -d postgres -c \
    "SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = '$DB_NAME' AND pid <> pg_backend_pid();" 2>/dev/null || true

# Restore from backup
if [[ "$BACKUP_FILE" == *.gz ]]; then
    gunzip -c "$BACKUP_FILE" | docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME"
else
    docker exec -i "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" < "$BACKUP_FILE"
fi

log_info "Database restored successfully from: $BACKUP_FILE"
log_info "Safety backup available at: $SAFETY_BACKUP"
