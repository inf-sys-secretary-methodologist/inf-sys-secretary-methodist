FROM alpine:3.21

# Install required packages
RUN apk add --no-cache \
    bash \
    postgresql16-client \
    curl \
    gzip \
    tar \
    tzdata \
    dcron \
    # For notifications and metrics
    jq \
    openssl \
    # For encryption (GPG)
    gnupg

# Install age encryption tool — version-pinned and SHA256-verified.
# Hashes computed from the official FiloSottile/age v1.2.0 release archives.
# --fail makes curl exit non-zero on HTTP errors (instead of writing an
# error page), --retry survives transient network failures, and the
# sha256sum -c check fails the build on any tampered/truncated download.
RUN ARCH=$(uname -m) && \
    case "$ARCH" in \
        aarch64|arm64) AGE_ARCH="arm64"; AGE_SHA256="d25a81f3ac011884009d18362eeb8154ce1bca4d151834c35c718654bd6c6353" ;; \
        x86_64|amd64) AGE_ARCH="amd64"; AGE_SHA256="2ae71cb3ea761118937a944083f057cfd42f0ef11d197ce72fc2b8780d50c4ef" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    AGE_VERSION="v1.2.0" && \
    curl -fsSL --retry 3 --retry-delay 2 --retry-connrefused \
        "https://github.com/FiloSottile/age/releases/download/${AGE_VERSION}/age-${AGE_VERSION}-linux-${AGE_ARCH}.tar.gz" \
        -o /tmp/age.tar.gz && \
    echo "${AGE_SHA256}  /tmp/age.tar.gz" | sha256sum -c - && \
    tar -xz -C /tmp -f /tmp/age.tar.gz && \
    mv /tmp/age/age /usr/local/bin/age && \
    mv /tmp/age/age-keygen /usr/local/bin/age-keygen && \
    chmod +x /usr/local/bin/age /usr/local/bin/age-keygen && \
    rm -rf /tmp/age /tmp/age.tar.gz

# Install MinIO Client (mc) - detect architecture
RUN ARCH=$(uname -m) && \
    case "$ARCH" in \
        aarch64|arm64) MC_ARCH="arm64" ;; \
        x86_64|amd64) MC_ARCH="amd64" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    curl -fsSL --retry 3 --retry-delay 2 --retry-connrefused \
        "https://dl.min.io/client/mc/release/linux-${MC_ARCH}/mc" -o /usr/local/bin/mc && \
    chmod +x /usr/local/bin/mc

# Create directories
RUN mkdir -p \
    /backups/postgres \
    /backups/minio \
    /scripts \
    /var/lib/node_exporter/textfile_collector

# Copy scripts
COPY scripts/ /scripts/
RUN chmod +x /scripts/*.sh

# Create entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set timezone
ENV TZ=Europe/Moscow

# Default environment variables
ENV METRICS_ENABLED=true \
    NOTIFY_ON_FAILURE=true \
    NOTIFY_ON_SUCCESS=false \
    BACKUP_ENCRYPTION_ENABLED=false \
    BACKUP_ENCRYPTION_TYPE=age

ENTRYPOINT ["/entrypoint.sh"]
