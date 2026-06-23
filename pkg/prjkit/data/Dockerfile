FROM alpine:latest AS first

# This path matches project path used in skw-dki-golang image.
RUN mkdir -p /vrinf/project

ENV VRIT_SVC_NAME="first"
ARG NAME="first"
RUN echo "first $NAME"
WORKDIR /vrinf/project
ENTRYPOINT ["echo", "first"]
CMD ["image"]

# Arguments expected in all Dockerfiles.
ARG VRIT_BUILD_DATE="0001-01-01T00:00:00Z"
ARG VRIT_CC_TAG="unknown"
ARG VRIT_SCM_REPO="unknown"
ARG VRIT_SCM_HASH="0000000"
ARG VRIT_SCM_REV="v0.0.0"

# Labels expected to be set in all images.
LABEL dev.vrit.build.date="$VRIT_BUILD_DATE" \
      dev.vrit.cc.build.tag="$VRIT_CC_TAG" \
      dev.vrit.scm.repo="$VRIT_SCM_REPO" \
      dev.vrit.scm.hash="$VRIT_SCM_HASH" \
      dev.vrit.scm.rev="$VRIT_SCM_REV" \
      dev.vrit.svc.name="first"

# Environemnt variables expected in all images.
ENV VRIT_BUILD_DATE="$VRIT_BUILD_DATE" \
    VRIT_CC_TAG="$VRIT_CC_TAG" \
    VRIT_SCM_REPO="$VRIT_SCM_REPO" \
    VRIT_SCM_HASH="$VRIT_SCM_HASH" \
    VRIT_SCM_REV="$VRIT_SCM_REV"

FROM first AS second

ENV VRIT_SVC_NAME="second"
LABEL dev.vrit.svc.name="second"

ARG NAME="second"
RUN echo "second $NAME"
ENTRYPOINT ["echo", "second"]
CMD ["image"]

FROM second AS third

ENV VRIT_SVC_NAME="third"
LABEL dev.vrit.svc.name="third"

ARG NAME="third"
RUN echo "second $NAME"
ENTRYPOINT ["echo", "third"]
CMD ["image"]

