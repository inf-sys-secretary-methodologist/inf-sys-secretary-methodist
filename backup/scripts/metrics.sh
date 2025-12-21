#!/bin/bash
# Prometheus metrics exporter for backup system
# Uses textfile collector pattern for node_exporter
# Usage: ./metrics.sh <action> [args]
#   record_backup postgres|minio <status> <duration_seconds> <size_bytes>
#   record_sync <status> <duration_seconds>
#   export (writes metrics to file)

set -euo pipefail

METRICS_DIR="${METRICS_DIR:-/var/lib/node_exporter/textfile_collector}"
METRICS_FILE="${METRICS_DIR}/backup_metrics.prom"
STATE_FILE="${METRICS_DIR}/.backup_state.json"
SERVER_ID="${SERVER_ID:-production}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [metrics] $1"
}

# Ensure metrics directory exists
ensure_dir() {
    if [[ ! -d "${METRICS_DIR}" ]]; then
        mkdir -p "${METRICS_DIR}"
        log "Created metrics directory: ${METRICS_DIR}"
    fi
}

# Initialize state file if not exists
init_state() {
    ensure_dir
    if [[ ! -f "${STATE_FILE}" ]]; then
        echo '{}' > "${STATE_FILE}"
    fi
}

# Get current timestamp
get_timestamp() {
    date +%s
}

# Record backup result to state
record_backup() {
    local backup_type="$1"  # postgres or minio
    local status="$2"       # success or failure
    local duration="$3"     # in seconds
    local size="${4:-0}"    # in bytes

    init_state

    local timestamp=$(get_timestamp)
    local success=$([[ "${status}" == "success" ]] && echo 1 || echo 0)

    # Read current state and update
    local state=$(cat "${STATE_FILE}")

    state=$(echo "${state}" | jq --arg type "${backup_type}" \
        --argjson timestamp "${timestamp}" \
        --argjson success "${success}" \
        --argjson duration "${duration}" \
        --argjson size "${size}" \
        '.[$type] = {
            "last_timestamp": $timestamp,
            "last_success": $success,
            "last_duration": $duration,
            "last_size": $size,
            "total_count": ((.[$type].total_count // 0) + 1),
            "success_count": ((.[$type].success_count // 0) + (if $success == 1 then 1 else 0 end)),
            "failure_count": ((.[$type].failure_count // 0) + (if $success == 0 then 1 else 0 end))
        }')

    # Update last success timestamp if successful
    if [[ "${status}" == "success" ]]; then
        state=$(echo "${state}" | jq --arg type "${backup_type}" \
            --argjson timestamp "${timestamp}" \
            '.[$type].last_success_timestamp = $timestamp')
    fi

    echo "${state}" > "${STATE_FILE}"
    log "Recorded ${backup_type} backup: status=${status}, duration=${duration}s, size=${size} bytes"

    # Export metrics after recording
    export_metrics
}

# Record remote sync result
record_sync() {
    local status="$1"       # success or failure
    local duration="$2"     # in seconds

    init_state

    local timestamp=$(get_timestamp)
    local success=$([[ "${status}" == "success" ]] && echo 1 || echo 0)

    local state=$(cat "${STATE_FILE}")

    state=$(echo "${state}" | jq \
        --argjson timestamp "${timestamp}" \
        --argjson success "${success}" \
        --argjson duration "${duration}" \
        '.remote_sync = {
            "last_timestamp": $timestamp,
            "last_success": $success,
            "last_duration": $duration,
            "total_count": ((.remote_sync.total_count // 0) + 1),
            "success_count": ((.remote_sync.success_count // 0) + (if $success == 1 then 1 else 0 end)),
            "failure_count": ((.remote_sync.failure_count // 0) + (if $success == 0 then 1 else 0 end))
        }')

    if [[ "${status}" == "success" ]]; then
        state=$(echo "${state}" | jq \
            --argjson timestamp "${timestamp}" \
            '.remote_sync.last_success_timestamp = $timestamp')
    fi

    echo "${state}" > "${STATE_FILE}"
    log "Recorded remote sync: status=${status}, duration=${duration}s"

    export_metrics
}

# Export metrics to Prometheus textfile format
export_metrics() {
    ensure_dir
    init_state

    local state=$(cat "${STATE_FILE}")
    local temp_file=$(mktemp)
    local current_time=$(get_timestamp)

    cat > "${temp_file}" << EOF
# HELP backup_last_run_timestamp_seconds Unix timestamp of the last backup run
# TYPE backup_last_run_timestamp_seconds gauge
EOF

    # Export postgres metrics
    if echo "${state}" | jq -e '.postgres' > /dev/null 2>&1; then
        local pg_timestamp=$(echo "${state}" | jq -r '.postgres.last_timestamp // 0')
        local pg_success=$(echo "${state}" | jq -r '.postgres.last_success // 0')
        local pg_duration=$(echo "${state}" | jq -r '.postgres.last_duration // 0')
        local pg_size=$(echo "${state}" | jq -r '.postgres.last_size // 0')
        local pg_success_ts=$(echo "${state}" | jq -r '.postgres.last_success_timestamp // 0')
        local pg_total=$(echo "${state}" | jq -r '.postgres.total_count // 0')
        local pg_success_count=$(echo "${state}" | jq -r '.postgres.success_count // 0')
        local pg_failure_count=$(echo "${state}" | jq -r '.postgres.failure_count // 0')

        cat >> "${temp_file}" << EOF
backup_last_run_timestamp_seconds{server_id="${SERVER_ID}",type="postgres"} ${pg_timestamp}

# HELP backup_last_success_timestamp_seconds Unix timestamp of the last successful backup
# TYPE backup_last_success_timestamp_seconds gauge
backup_last_success_timestamp_seconds{server_id="${SERVER_ID}",type="postgres"} ${pg_success_ts}

# HELP backup_last_run_success Whether the last backup was successful (1) or failed (0)
# TYPE backup_last_run_success gauge
backup_last_run_success{server_id="${SERVER_ID}",type="postgres"} ${pg_success}

# HELP backup_duration_seconds Duration of the last backup in seconds
# TYPE backup_duration_seconds gauge
backup_duration_seconds{server_id="${SERVER_ID}",type="postgres"} ${pg_duration}

# HELP backup_size_bytes Size of the last backup in bytes
# TYPE backup_size_bytes gauge
backup_size_bytes{server_id="${SERVER_ID}",type="postgres"} ${pg_size}

# HELP backup_total_count Total number of backup attempts
# TYPE backup_total_count counter
backup_total_count{server_id="${SERVER_ID}",type="postgres"} ${pg_total}

# HELP backup_success_count Total number of successful backups
# TYPE backup_success_count counter
backup_success_count{server_id="${SERVER_ID}",type="postgres"} ${pg_success_count}

# HELP backup_failure_count Total number of failed backups
# TYPE backup_failure_count counter
backup_failure_count{server_id="${SERVER_ID}",type="postgres"} ${pg_failure_count}

# HELP backup_age_seconds Time since last successful backup in seconds
# TYPE backup_age_seconds gauge
backup_age_seconds{server_id="${SERVER_ID}",type="postgres"} $((current_time - pg_success_ts))
EOF
    fi

    # Export minio metrics
    if echo "${state}" | jq -e '.minio' > /dev/null 2>&1; then
        local minio_timestamp=$(echo "${state}" | jq -r '.minio.last_timestamp // 0')
        local minio_success=$(echo "${state}" | jq -r '.minio.last_success // 0')
        local minio_duration=$(echo "${state}" | jq -r '.minio.last_duration // 0')
        local minio_size=$(echo "${state}" | jq -r '.minio.last_size // 0')
        local minio_success_ts=$(echo "${state}" | jq -r '.minio.last_success_timestamp // 0')
        local minio_total=$(echo "${state}" | jq -r '.minio.total_count // 0')
        local minio_success_count=$(echo "${state}" | jq -r '.minio.success_count // 0')
        local minio_failure_count=$(echo "${state}" | jq -r '.minio.failure_count // 0')

        cat >> "${temp_file}" << EOF

backup_last_run_timestamp_seconds{server_id="${SERVER_ID}",type="minio"} ${minio_timestamp}
backup_last_success_timestamp_seconds{server_id="${SERVER_ID}",type="minio"} ${minio_success_ts}
backup_last_run_success{server_id="${SERVER_ID}",type="minio"} ${minio_success}
backup_duration_seconds{server_id="${SERVER_ID}",type="minio"} ${minio_duration}
backup_size_bytes{server_id="${SERVER_ID}",type="minio"} ${minio_size}
backup_total_count{server_id="${SERVER_ID}",type="minio"} ${minio_total}
backup_success_count{server_id="${SERVER_ID}",type="minio"} ${minio_success_count}
backup_failure_count{server_id="${SERVER_ID}",type="minio"} ${minio_failure_count}
backup_age_seconds{server_id="${SERVER_ID}",type="minio"} $((current_time - minio_success_ts))
EOF
    fi

    # Export remote sync metrics
    if echo "${state}" | jq -e '.remote_sync' > /dev/null 2>&1; then
        local sync_timestamp=$(echo "${state}" | jq -r '.remote_sync.last_timestamp // 0')
        local sync_success=$(echo "${state}" | jq -r '.remote_sync.last_success // 0')
        local sync_duration=$(echo "${state}" | jq -r '.remote_sync.last_duration // 0')
        local sync_success_ts=$(echo "${state}" | jq -r '.remote_sync.last_success_timestamp // 0')
        local sync_total=$(echo "${state}" | jq -r '.remote_sync.total_count // 0')
        local sync_success_count=$(echo "${state}" | jq -r '.remote_sync.success_count // 0')
        local sync_failure_count=$(echo "${state}" | jq -r '.remote_sync.failure_count // 0')

        cat >> "${temp_file}" << EOF

# HELP backup_remote_sync_last_run_timestamp_seconds Unix timestamp of last remote sync
# TYPE backup_remote_sync_last_run_timestamp_seconds gauge
backup_remote_sync_last_run_timestamp_seconds{server_id="${SERVER_ID}"} ${sync_timestamp}

# HELP backup_remote_sync_last_success_timestamp_seconds Unix timestamp of last successful remote sync
# TYPE backup_remote_sync_last_success_timestamp_seconds gauge
backup_remote_sync_last_success_timestamp_seconds{server_id="${SERVER_ID}"} ${sync_success_ts}

# HELP backup_remote_sync_last_run_success Whether the last sync was successful
# TYPE backup_remote_sync_last_run_success gauge
backup_remote_sync_last_run_success{server_id="${SERVER_ID}"} ${sync_success}

# HELP backup_remote_sync_duration_seconds Duration of last remote sync
# TYPE backup_remote_sync_duration_seconds gauge
backup_remote_sync_duration_seconds{server_id="${SERVER_ID}"} ${sync_duration}

# HELP backup_remote_sync_total_count Total sync attempts
# TYPE backup_remote_sync_total_count counter
backup_remote_sync_total_count{server_id="${SERVER_ID}"} ${sync_total}

# HELP backup_remote_sync_success_count Total successful syncs
# TYPE backup_remote_sync_success_count counter
backup_remote_sync_success_count{server_id="${SERVER_ID}"} ${sync_success_count}

# HELP backup_remote_sync_failure_count Total failed syncs
# TYPE backup_remote_sync_failure_count counter
backup_remote_sync_failure_count{server_id="${SERVER_ID}"} ${sync_failure_count}
EOF
    fi

    # Atomically move temp file to metrics file
    mv "${temp_file}" "${METRICS_FILE}"
    log "Metrics exported to ${METRICS_FILE}"
}

# Main dispatch
case "${1:-}" in
    record_backup)
        shift
        record_backup "$@"
        ;;
    record_sync)
        shift
        record_sync "$@"
        ;;
    export)
        export_metrics
        ;;
    *)
        echo "Usage: $0 <action> [args]"
        echo "Actions:"
        echo "  record_backup <type> <status> <duration> [size]"
        echo "  record_sync <status> <duration>"
        echo "  export"
        exit 1
        ;;
esac
