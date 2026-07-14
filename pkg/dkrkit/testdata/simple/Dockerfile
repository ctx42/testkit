# ==============================================================================
# Simple test image.
# ==============================================================================
ARG OCI_IMAGE_BASE_NAME
FROM $OCI_IMAGE_BASE_NAME

# Arguments expected in all Dockerfiles.
ARG OCI_IMAGE_BASE_NAME
ARG OCI_IMAGE_CREATED="0001-01-01T00:00:00Z"
ARG OCI_IMAGE_AUTHORS="unknown"
ARG OCI_IMAGE_REF_NAME="unknown"
ARG OCI_IMAGE_SOURCE="unknown"
ARG OCI_IMAGE_REVISION="0000000"
ARG OCI_IMAGE_VERSION="v0.0.0"

# Labels expected to be set in all images.
LABEL org.opencontainers.image.created="$OCI_IMAGE_CREATED" \
      org.opencontainers.image.authors="$OCI_IMAGE_AUTHORS" \
      org.opencontainers.image.ref.name="$OCI_IMAGE_REF_NAME" \
      org.opencontainers.image.base.name="$OCI_IMAGE_BASE_NAME" \
      org.opencontainers.image.source="$OCI_IMAGE_SOURCE" \
      org.opencontainers.image.revision="$OCI_IMAGE_REVISION" \
      org.opencontainers.image.version="$OCI_IMAGE_VERSION"

# Environemnt variables expected in all images.
ENV OCI_IMAGE_CREATED="$OCI_IMAGE_CREATED" \
    OCI_IMAGE_AUTHORS="$OCI_IMAGE_AUTHORS" \
    OCI_IMAGE_REF_NAME="$OCI_IMAGE_REF_NAME" \
    OCI_IMAGE_BASE_NAME="$OCI_IMAGE_BASE_NAME" \
    OCI_IMAGE_SOURCE="$OCI_IMAGE_SOURCE" \
    OCI_IMAGE_REVISION="$OCI_IMAGE_REVISION" \
    OCI_IMAGE_VERSION="$OCI_IMAGE_VERSION"

# Test environment variables.
ENV C42_TEST_EMPTY="" \
    C42_TEST_VALUE="value"

# Test labels.
LABEL com.ctx42.test.empty="" \
      com.ctx42.test.value="value"

COPY <<-EOF /file0.txt
file0.txt content
EOF

COPY <<-EOF /entrypoint.sh
#!/usr/bin/env sh

if [ \$# -eq 0 ]; then
    echo "hello"
    exit 0
fi

eval "\$@"
EOF

RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
