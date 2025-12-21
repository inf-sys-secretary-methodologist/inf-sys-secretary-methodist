#!/bin/bash
# Notification script for backup system
# Supports: Telegram, Webhook, Email (SMTP)
# Usage: ./notify.sh <level> <title> <message>
#   level: success, warning, error, info

set -euo pipefail

LEVEL="${1:-info}"
TITLE="${2:-Backup Notification}"
MESSAGE="${3:-No message provided}"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
HOSTNAME=$(hostname 2>/dev/null || echo "backup-service")
SERVER_ID="${SERVER_ID:-production}"

# Notification settings from environment
NOTIFY_TELEGRAM_ENABLED="${NOTIFY_TELEGRAM_ENABLED:-false}"
NOTIFY_TELEGRAM_BOT_TOKEN="${NOTIFY_TELEGRAM_BOT_TOKEN:-}"
NOTIFY_TELEGRAM_CHAT_ID="${NOTIFY_TELEGRAM_CHAT_ID:-}"

NOTIFY_WEBHOOK_ENABLED="${NOTIFY_WEBHOOK_ENABLED:-false}"
NOTIFY_WEBHOOK_URL="${NOTIFY_WEBHOOK_URL:-}"
NOTIFY_WEBHOOK_SECRET="${NOTIFY_WEBHOOK_SECRET:-}"

NOTIFY_EMAIL_ENABLED="${NOTIFY_EMAIL_ENABLED:-false}"
NOTIFY_EMAIL_SMTP_HOST="${NOTIFY_EMAIL_SMTP_HOST:-}"
NOTIFY_EMAIL_SMTP_PORT="${NOTIFY_EMAIL_SMTP_PORT:-587}"
NOTIFY_EMAIL_FROM="${NOTIFY_EMAIL_FROM:-}"
NOTIFY_EMAIL_TO="${NOTIFY_EMAIL_TO:-}"
NOTIFY_EMAIL_USER="${NOTIFY_EMAIL_USER:-}"
NOTIFY_EMAIL_PASSWORD="${NOTIFY_EMAIL_PASSWORD:-}"

# Only notify on these levels (comma-separated, e.g., "error,warning")
NOTIFY_LEVELS="${NOTIFY_LEVELS:-error,warning,success}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [notify] $1"
}

# Check if current level should trigger notification
should_notify() {
    echo "${NOTIFY_LEVELS}" | tr ',' '\n' | grep -qx "${LEVEL}"
}

# Get emoji based on level
get_emoji() {
    case "${LEVEL}" in
        success) echo "✅" ;;
        warning) echo "⚠️" ;;
        error)   echo "🚨" ;;
        info)    echo "ℹ️" ;;
        *)       echo "📋" ;;
    esac
}

# Send Telegram notification
send_telegram() {
    if [[ "${NOTIFY_TELEGRAM_ENABLED}" != "true" ]]; then
        return 0
    fi

    if [[ -z "${NOTIFY_TELEGRAM_BOT_TOKEN}" ]] || [[ -z "${NOTIFY_TELEGRAM_CHAT_ID}" ]]; then
        log "Telegram: Missing BOT_TOKEN or CHAT_ID, skipping"
        return 0
    fi

    local emoji=$(get_emoji)
    local text="<b>${emoji} ${TITLE}</b>

<b>Server:</b> ${SERVER_ID}
<b>Level:</b> ${LEVEL^^}
<b>Time:</b> ${TIMESTAMP}

<pre>${MESSAGE}</pre>"

    log "Telegram: Sending notification..."

    local response
    response=$(curl -s -X POST \
        "https://api.telegram.org/bot${NOTIFY_TELEGRAM_BOT_TOKEN}/sendMessage" \
        -H "Content-Type: application/json" \
        -d "{
            \"chat_id\": \"${NOTIFY_TELEGRAM_CHAT_ID}\",
            \"text\": $(echo "${text}" | jq -Rs .),
            \"parse_mode\": \"HTML\",
            \"disable_notification\": $([ "${LEVEL}" = "info" ] && echo "true" || echo "false")
        }" 2>&1) || {
        log "Telegram: Failed to send notification"
        return 1
    }

    if echo "${response}" | jq -e '.ok == true' > /dev/null 2>&1; then
        log "Telegram: Notification sent successfully"
    else
        log "Telegram: Failed - $(echo "${response}" | jq -r '.description // "Unknown error"')"
        return 1
    fi
}

# Send Webhook notification
send_webhook() {
    if [[ "${NOTIFY_WEBHOOK_ENABLED}" != "true" ]]; then
        return 0
    fi

    if [[ -z "${NOTIFY_WEBHOOK_URL}" ]]; then
        log "Webhook: Missing URL, skipping"
        return 0
    fi

    log "Webhook: Sending notification to ${NOTIFY_WEBHOOK_URL}..."

    local payload="{
        \"level\": \"${LEVEL}\",
        \"title\": \"${TITLE}\",
        \"message\": \"${MESSAGE}\",
        \"timestamp\": \"${TIMESTAMP}\",
        \"server_id\": \"${SERVER_ID}\",
        \"hostname\": \"${HOSTNAME}\"
    }"

    local headers="-H \"Content-Type: application/json\""
    if [[ -n "${NOTIFY_WEBHOOK_SECRET}" ]]; then
        local signature=$(echo -n "${payload}" | openssl dgst -sha256 -hmac "${NOTIFY_WEBHOOK_SECRET}" | cut -d' ' -f2)
        headers="${headers} -H \"X-Webhook-Signature: sha256=${signature}\""
    fi

    local response
    response=$(curl -s -X POST "${NOTIFY_WEBHOOK_URL}" \
        -H "Content-Type: application/json" \
        ${NOTIFY_WEBHOOK_SECRET:+-H "X-Webhook-Signature: sha256=$(echo -n "${payload}" | openssl dgst -sha256 -hmac "${NOTIFY_WEBHOOK_SECRET}" | cut -d' ' -f2)"} \
        -d "${payload}" 2>&1) || {
        log "Webhook: Failed to send notification"
        return 1
    }

    log "Webhook: Notification sent"
}

# Send Email notification via SMTP
send_email() {
    if [[ "${NOTIFY_EMAIL_ENABLED}" != "true" ]]; then
        return 0
    fi

    if [[ -z "${NOTIFY_EMAIL_SMTP_HOST}" ]] || [[ -z "${NOTIFY_EMAIL_TO}" ]] || [[ -z "${NOTIFY_EMAIL_FROM}" ]]; then
        log "Email: Missing SMTP configuration, skipping"
        return 0
    fi

    local emoji=$(get_emoji)
    local subject="${emoji} [${LEVEL^^}] ${TITLE} - ${SERVER_ID}"

    local body="Backup Notification
=====================================

Level:     ${LEVEL^^}
Server:    ${SERVER_ID}
Hostname:  ${HOSTNAME}
Time:      ${TIMESTAMP}

Message:
${MESSAGE}

---
Sent by backup-service
"

    log "Email: Sending notification to ${NOTIFY_EMAIL_TO}..."

    # Using curl to send email via SMTP
    local auth_params=""
    if [[ -n "${NOTIFY_EMAIL_USER}" ]] && [[ -n "${NOTIFY_EMAIL_PASSWORD}" ]]; then
        auth_params="--user \"${NOTIFY_EMAIL_USER}:${NOTIFY_EMAIL_PASSWORD}\""
    fi

    # Create email file
    local email_file=$(mktemp)
    cat > "${email_file}" << EOF
From: ${NOTIFY_EMAIL_FROM}
To: ${NOTIFY_EMAIL_TO}
Subject: ${subject}
Content-Type: text/plain; charset=UTF-8

${body}
EOF

    curl -s --ssl-reqd \
        --url "smtp://${NOTIFY_EMAIL_SMTP_HOST}:${NOTIFY_EMAIL_SMTP_PORT}" \
        --mail-from "${NOTIFY_EMAIL_FROM}" \
        --mail-rcpt "${NOTIFY_EMAIL_TO}" \
        ${NOTIFY_EMAIL_USER:+--user "${NOTIFY_EMAIL_USER}:${NOTIFY_EMAIL_PASSWORD}"} \
        --upload-file "${email_file}" 2>&1 && {
        log "Email: Notification sent successfully"
        rm -f "${email_file}"
    } || {
        log "Email: Failed to send notification"
        rm -f "${email_file}"
        return 1
    }
}

# Main
main() {
    if ! should_notify; then
        log "Level '${LEVEL}' not in NOTIFY_LEVELS (${NOTIFY_LEVELS}), skipping notifications"
        exit 0
    fi

    local errors=0

    # Send to all configured channels
    send_telegram || ((errors++)) || true
    send_webhook || ((errors++)) || true
    send_email || ((errors++)) || true

    if [[ ${errors} -gt 0 ]]; then
        log "Some notifications failed (${errors} error(s))"
        exit 1
    fi

    log "All notifications sent successfully"
}

main
