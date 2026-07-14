FROM alpine:latest AS first

# This path matches project path used in ctx42/dkigo image.
RUN mkdir -p /ctx42/project

ENV OCI_IMAGE_TITLE="first"
ARG NAME="first"
RUN echo "first $NAME"
WORKDIR /ctx42/project
ENTRYPOINT ["echo", "first"]
CMD ["image"]

# Arguments expected in all Dockerfiles.
ARG C42_BUILD_DATE="0001-01-01T00:00:00Z"
ARG C42_CCID="unknown"
ARG C42_SCM_REPO="unknown"
ARG C42_SCM_HASH="0000000"
ARG C42_SCM_REV="v0.0.0"

# Labels expected to be set in all images.
LABEL org.opencontainers.image.created="$C42_BUILD_DATE" \
      org.opencontainers.image.ref.name="$C42_CCID" \
      org.opencontainers.image.source="$C42_SCM_REPO" \
      org.opencontainers.image.revision="$C42_SCM_HASH" \
      org.opencontainers.image.version="$C42_SCM_REV" \
      org.opencontainers.image.title="first"

# Environemnt variables expected in all images.
ENV OCI_IMAGE_CREATED="$C42_BUILD_DATE" \
    OCI_IMAGE_REF_NAME="$C42_CCID" \
    OCI_IMAGE_SOURCE="$C42_SCM_REPO" \
    OCI_IMAGE_REVISION="$C42_SCM_HASH" \
    OCI_IMAGE_VERSION="$C42_SCM_REV"

FROM first AS second

ENV OCI_IMAGE_TITLE="second"
LABEL org.opencontainers.image.title="second"

ARG NAME="second"
RUN echo "second $NAME"
ENTRYPOINT ["echo", "second"]
CMD ["image"]

FROM second AS third

ENV OCI_IMAGE_TITLE="third"
LABEL org.opencontainers.image.title="third"

ARG NAME="third"
RUN echo "second $NAME"
ENTRYPOINT ["echo", "third"]
CMD ["image"]

