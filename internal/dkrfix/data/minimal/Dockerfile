# ==============================================================================
# Minimal test image.
#
# The smallest image the fixtures offer, returned by dkrfix.Minimal: it takes
# the base image from the C42_BLD_IMG_BASE build argument and adds only what
# the tests inspect - test labels and environment variables (from
# C42_TST_NAME), a /file0.txt file to copy out of a container, and an
# /entrypoint.sh which prints "hello" when run without arguments and evaluates
# them otherwise, so a container can be started both as a short-lived and a
# long-running one.
#
# Unlike data/simple/Dockerfile it carries no OCI build metadata, which keeps
# builds fast and the label and environment sets small enough to assert on in
# full.
# ==============================================================================

ARG C42_BLD_IMG_BASE
FROM $C42_BLD_IMG_BASE

# Added to test values of labels and environmental variables.
ARG C42_TST_NAME="unknown"
LABEL com.ctx42.test.empty="" \
    com.ctx42.test.name="$C42_TST_NAME" \
    com.ctx42.test.value="value"
ENV C42_TST_EMPTY="" \
    C42_TST_NAME="$C42_TST_NAME" \
    C42_TST_VALUE="value"

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
