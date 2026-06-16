ARG OCI_IMAGE_BASE_NAME
FROM $OCI_IMAGE_BASE_NAME

ARG OCI_IMAGE_TITLE
ARG OCI_IMAGE_BASE_NAME

LABEL org.opencontainers.image.created="2000-01-02T03:04:05Z" \
      org.opencontainers.image.base.name="$OCI_IMAGE_BASE_NAME" \
      org.opencontainers.image.title=$OCI_IMAGE_TITLE \
      com.ctx42.test.empty=""

ENV OCI_IMAGE_CREATED="2000-01-02T03:04:05Z" \
    OCI_IMAGE_BASE_NAME="$OCI_IMAGE_BASE_NAME" \
    OCI_IMAGE_TITLE=$OCI_IMAGE_TITLE \
    CTX42_TEST_EMPTY=""

COPY <<-EOF /file0.txt
file0
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
