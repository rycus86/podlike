FROM golang:1.10 as builder

ARG CC=""
ARG CC_PKG=""
ARG CC_GOARCH=""

ADD . /go/src/github.com/rycus86/podlike
WORKDIR /go/src/github.com/rycus86/podlike

RUN if [ -n "$CC_PKG" ]; then \
      apt-get update && apt-get install -y $CC_PKG; \
    fi \
    && export CC=$CC \
    && export GOOS=linux \
    && export GOARCH=$CC_GOARCH \
    && export CGO_ENABLED=0 \
    && go build -o /var/out/main -v ./cmd/podlike \
    && go build -o /var/out/hc   -v ./cmd/healthcheck

FROM scratch

ARG VERSION="dev"
ARG BUILD_ARCH="unknown"
ARG GIT_COMMIT="unknown"
ARG BUILD_TIMESTAMP="unknown"

ENV VERSION="$VERSION"
ENV BUILD_ARCH="$BUILD_ARCH"
ENV GIT_COMMIT="$GIT_COMMIT"
ENV BUILD_TIMESTAMP="$BUILD_TIMESTAMP"

LABEL maintainer="Viktor Adam <rycus86@gmail.com>"

LABEL com.github.rycus86.podlike.version="$VERSION"
LABEL com.github.rycus86.podlike.commit="$GIT_COMMIT"

COPY --from=builder /var/out/main  /podlike

HEALTHCHECK --interval=2s --timeout=3s --retries=5 CMD [ "/podlike", "healthcheck" ]

ENTRYPOINT [ "/podlike" ]
