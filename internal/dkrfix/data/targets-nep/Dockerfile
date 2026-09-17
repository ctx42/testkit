# ==============================================================================
# Example multi-target project image without an entrypoint (NEP).
#
# The Dockerfile returned by dkrfix.TargetsNEP: it takes the base image from
# the C42_BLD_IMG_BASE build argument and defines three chained targets -
# first, second and third - so a test can build a selected target and tell the
# resulting images apart. Every target declares the build arguments expected in
# all ctx42 Dockerfiles, stamps the OCI image metadata labels and C42_*
# environment variables from them unchanged, and re-stamps C42_BLD_IMG_TARGET
# and the com.ctx42.image.target label with its own name. C42_SCM_REV arrives
# ready to stamp and no ctx42 Dockerfile adds anything to it, the build target
# included.
#
# Unlike data/targets/Dockerfile no target sets an ENTRYPOINT or a CMD, which
# is what makes this fixture useful - it covers the images a run helper cannot
# start without an explicit command. The two files are otherwise identical and
# must be kept in lockstep.
# ==============================================================================
ARG C42_BLD_IMG_BASE="busybox:1.38-uclibc"
FROM $C42_BLD_IMG_BASE AS first

# This path matches project path used in ctx42/dkigo image.
RUN mkdir -p /ctx42/project
WORKDIR /ctx42/project

ARG NAME="first"
RUN echo "first $NAME"

# Arguments expected in all Dockerfiles.
ARG C42_BLD_DATE="0001-01-01T00:00:00Z"
ARG C42_PRJ_NAME="unknown"
ARG C42_SCM_REPO="unknown"
ARG C42_SCM_HASH="0000000"
ARG C42_SCM_REV="v0.0.0"
ARG C42_SCM_STATE="unknown"

# Environment variables expected in all images. Set before the labels, so the
# labels can be stamped from the environment and the two cannot drift apart.
ENV C42_BLD_DATE="$C42_BLD_DATE" \
    C42_PRJ_NAME="$C42_PRJ_NAME" \
    C42_SCM_HASH="$C42_SCM_HASH" \
    C42_SCM_REV="$C42_SCM_REV" \
    C42_SCM_STATE="$C42_SCM_STATE" \
    C42_BLD_IMG_TARGET="first"

# Labels expected to be set in all images. The chained targets below inherit
# the OCI ones unchanged and re-stamp com.ctx42.image.target with their own
# name, so only the target tells the images apart. The OCI Image Spec has no
# annotation for a build target, so the key is a ctx42 one.
LABEL org.opencontainers.image.created="$C42_BLD_DATE" \
      org.opencontainers.image.revision="$C42_SCM_HASH" \
      org.opencontainers.image.version="$C42_SCM_REV" \
      org.opencontainers.image.source="$C42_SCM_REPO" \
      com.ctx42.image.target="first"

FROM first AS second

ARG NAME="second"
RUN echo "second $NAME"

ENV C42_BLD_IMG_TARGET="second"
LABEL com.ctx42.image.target="second"

FROM second AS third

ARG NAME="third"
RUN echo "third $NAME"

ENV C42_BLD_IMG_TARGET="third"
LABEL com.ctx42.image.target="third"
