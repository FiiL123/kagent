### STAGE 1: base image
ARG BASE_IMAGE_REGISTRY=cgr.dev
ARG BUILDPLATFORM
FROM --platform=$BUILDPLATFORM $BASE_IMAGE_REGISTRY/chainguard/go:latest AS builder
ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILDPLATFORM

# Which package to build (e.g. core/cmd/controller/main.go, adk/cmd/main.go)
ARG BUILD_PACKAGE=core/cmd/controller/main.go

WORKDIR /workspace
# Copy the Go module manifests
COPY go.mod go.sum .
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN --mount=type=cache,target=/root/go/pkg/mod,rw      \
    --mount=type=cache,target=/root/.cache/go-build,rw \
     go mod download
# Copy all source
COPY api/ api/
COPY core/ core/
COPY adk/ adk/

# Build
ARG LDFLAGS
RUN --mount=type=cache,target=/root/go/pkg/mod,rw             \
    --mount=type=cache,target=/root/.cache/go-build,rw        \
    echo "Building on $BUILDPLATFORM -> linux/$TARGETARCH" && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -ldflags "$LDFLAGS" -o /app "$BUILD_PACKAGE"

### STAGE 2: install srt
FROM node:18-alpine AS srt-installer
RUN --mount=type=cache,target=/var/cache/apk,rw \
    apk add --no-cache git bubblewrap socat ripgrep && \
    mkdir -p /opt/sandbox-runtime && \
    git clone https://github.com/anthropic-experimental/sandbox-runtime.git /opt/sandbox-runtime && \
    cd /opt/sandbox-runtime && \
    git checkout -q ef4afdef4d711ba21a507d7f7369e305f7d3dbfa && \
    npm install && \
    npm run build && \
    npm prune --omit=dev

# Create srt wrapper script in node_modules/.bin
RUN mkdir -p /opt/sandbox-runtime/node_modules/.bin && \
    echo '#!/bin/sh' > /opt/sandbox-runtime/node_modules/.bin/srt && \
    echo 'exec node /opt/sandbox-runtime/dist/cli.js "$@"' >> /opt/sandbox-runtime/node_modules/.bin/srt && \
    chmod +x /opt/sandbox-runtime/node_modules/.bin/srt

### STAGE 3: final image
FROM alpine:latest
ARG TARGETPLATFORM

# Install runtime dependencies for srt and ca-certificates
RUN --mount=type=cache,target=/var/cache/apk,rw \
    apk add --no-cache bubblewrap nodejs npm ca-certificates ripgrep socat bash curl && \
    addgroup -g 65532 nonroot && \
    adduser -D -u 65532 -G nonroot nonroot

# Copy the Go binary
COPY --from=builder /app /app

# Copy srt installation
COPY --from=srt-installer /opt/sandbox-runtime /opt/sandbox-runtime

ENV PATH="/opt/sandbox-runtime/node_modules/.bin:$PATH"

USER 65532:65532
ARG VERSION

LABEL org.opencontainers.image.source=https://github.com/kagent-dev/kagent
LABEL org.opencontainers.image.description="Kagent app is the UI and apiserver for running agents."
LABEL org.opencontainers.image.version="$VERSION"

ENTRYPOINT ["/app"]
