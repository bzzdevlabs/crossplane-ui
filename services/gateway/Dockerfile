# syntax=docker/dockerfile:1.7
#
# Build context: repository root. The gateway image embeds the built Vue SPA
# via //go:embed, so we need access to web/ui/ as well as pkg/ and
# services/gateway/ at build time.

# ---- UI builder ----
FROM node:22-alpine AS ui-builder

WORKDIR /src

# Pin the same pnpm version declared in web/ui/package.json → packageManager.
ENV PNPM_HOME=/pnpm
ENV PATH=$PNPM_HOME:$PATH
RUN corepack enable

COPY web/ui/package.json web/ui/pnpm-lock.yaml ./web/ui/
WORKDIR /src/web/ui
RUN --mount=type=cache,target=/pnpm/store \
    pnpm install --frozen-lockfile

COPY web/ui/ ./
RUN pnpm run build
# Produces /src/web/ui/dist, ingested by the Go builder below.

# ---- Go builder ----
FROM golang:1.26-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /src

COPY go.work go.work.sum* ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY services/gateway/go.mod services/gateway/go.sum* ./services/gateway/
COPY services/auth/go.mod services/auth/go.sum* ./services/auth/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd services/gateway && go mod download

COPY pkg ./pkg
COPY services/gateway ./services/gateway
# Overlay the built SPA into the //go:embed target. Must happen before the
# Go build so the compiler picks it up.
COPY --from=ui-builder /src/web/ui/dist \
     ./services/gateway/internal/webui/dist

ENV CGO_ENABLED=0
ENV GOFLAGS=-trimpath
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    cd services/gateway && go build \
        -ldflags "-s -w \
            -X github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/buildinfo.Version=${VERSION} \
            -X github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/buildinfo.Commit=${COMMIT} \
            -X github.com/bzzdevlabs/crossplane-ui/services/gateway/internal/buildinfo.Date=${BUILD_DATE}" \
        -o /out/gateway ./cmd

# ---- Runtime ----
# distroless/static: no shell, no package manager, nothing but libc and the
# certificate store. Non-root user UID/GID 65532 (nonroot).
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

LABEL org.opencontainers.image.title="crossplane-ui gateway"
LABEL org.opencontainers.image.source="https://github.com/bzzdevlabs/crossplane-ui"
LABEL org.opencontainers.image.licenses="Apache-2.0"

USER 65532:65532
WORKDIR /
COPY --from=builder /out/gateway /usr/local/bin/gateway

EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gateway"]
