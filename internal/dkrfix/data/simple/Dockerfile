# ==============================================================================
# Simple test image.
#
# The full-featured test image: it takes the base image from the
# C42_BLD_IMG_BASE build argument and adds everything the tests inspect - the
# OCI image metadata labels and C42_* environment variables derived from the
# C42_BLD_* and C42_SCM_* build arguments, test labels and environment
# variables (from C42_TST_NAME), a /file0.txt file to copy out of a container,
# and an /entrypoint.sh which prints "hello" when run without arguments and
# evaluates them otherwise, so a container can be started both as a
# short-lived and a long-running one.
#
# Returned by dkrfix.Simple, it is the fixture a test builds when it needs an
# image carrying the full build metadata, so it must build with no more than
# the base image argument - every other argument has a default, or is used
# only where an empty value is acceptable.
#
# Use data/minimal/Dockerfile instead when the build metadata is irrelevant
# and small, fully assertable label and environment sets matter more.
# ==============================================================================

ARG C42_BLD_IMG_BASE
FROM $C42_BLD_IMG_BASE

ARG C42_BLD_DATE="0001-01-01T00:00:00Z"
ARG C42_PRJ_NAME="unknown"
ARG C42_SCM_HASH="0000000"
ARG C42_SCM_REV="v0.0.0"
ARG C42_SCM_REPO="unknown"
ARG C42_SCM_STATE="unknown"

LABEL org.opencontainers.image.created="$C42_BLD_DATE" \
      org.opencontainers.image.revision="$C42_SCM_HASH" \
      org.opencontainers.image.version="$C42_SCM_REV" \
      org.opencontainers.image.source="$C42_SCM_REPO"

ENV C42_BLD_DATE="$C42_BLD_DATE" \
    C42_PRJ_NAME="$C42_PRJ_NAME" \
    C42_SCM_HASH="$C42_SCM_HASH" \
    C42_SCM_REV="$C42_SCM_REV" \
    C42_SCM_STATE="$C42_SCM_STATE"

# Added to test values of labels and environmental variables.
ARG C42_TST_NAME
LABEL com.ctx42.test.empty="" com.ctx42.test.name="$C42_TST_NAME"
ENV C42_TST_EMPTY="" C42_TST_NAME="$C42_TST_NAME"

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
